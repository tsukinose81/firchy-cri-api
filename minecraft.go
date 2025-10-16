package criapi

import (
	"context"
	"fmt"
	"time"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// MinecraftServerConfig contains configuration for a Minecraft server
type MinecraftServerConfig struct {
	// Pod configuration
	PodName      string
	Namespace    string
	UID          string
	
	// Container configuration
	ContainerName string
	Image         string
	
	// Minecraft server configuration
	ServerType   string // e.g., "PAPER", "VANILLA", "SPIGOT"
	EULA         bool
	
	// Port mapping
	ServerPort   int32
	HostPort     int32
	
	// Additional environment variables
	ExtraEnv     map[string]string
}

// DefaultMinecraftConfig returns a default Minecraft server configuration
func DefaultMinecraftConfig() *MinecraftServerConfig {
	return &MinecraftServerConfig{
		PodName:       "minecraft-pod",
		Namespace:     "default",
		UID:           "minecraft-123",
		ContainerName: "minecraft-server",
		Image:         "docker.io/itzg/minecraft-server:latest",
		ServerType:    "PAPER",
		EULA:          true,
		ServerPort:    25565,
		HostPort:      25565,
		ExtraEnv:      make(map[string]string),
	}
}

// Validate checks if the MinecraftServerConfig is valid
func (c *MinecraftServerConfig) Validate() error {
	if c.PodName == "" {
		return fmt.Errorf("pod name cannot be empty")
	}
	if c.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if c.UID == "" {
		return fmt.Errorf("UID cannot be empty")
	}
	if c.ContainerName == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	if c.Image == "" {
		return fmt.Errorf("image cannot be empty")
	}
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.ServerPort)
	}
	if c.HostPort <= 0 || c.HostPort > 65535 {
		return fmt.Errorf("host port must be between 1 and 65535, got %d", c.HostPort)
	}
	return nil
}

// MinecraftServer represents a running Minecraft server instance
type MinecraftServer struct {
	Client      *Client
	Config      *MinecraftServerConfig
	PodID       string
	ContainerID string
}

// StartMinecraftServer creates and starts a Minecraft server with Kata runtime
func (c *Client) StartMinecraftServer(ctx context.Context, config *MinecraftServerConfig) (*MinecraftServer, error) {
	if config == nil {
		config = DefaultMinecraftConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Pull the image first
	fmt.Printf("Pulling image: %s\n", config.Image)
	_, err := c.PullImage(ctx, config.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create port mappings
	portMappings := []*runtime.PortMapping{
		{
			Protocol:      runtime.Protocol_TCP,
			ContainerPort: config.ServerPort,
			HostPort:      config.HostPort,
		},
	}

	// Create pod config
	podConfig := &runtime.PodSandboxConfig{
		Metadata: &runtime.PodSandboxMetadata{
			Name:      config.PodName,
			Namespace: config.Namespace,
			Uid:       config.UID,
		},
		Annotations: map[string]string{
			"io.containerd.cri.runtime-handler": "kata",
		},
		PortMappings: portMappings,
		DnsConfig: &runtime.DNSConfig{
			Servers: []string{"8.8.8.8", "8.8.4.4"},
		},
	}

	// Create and run pod sandbox
	fmt.Printf("Creating pod sandbox: %s\n", config.PodName)
	podID, err := c.RunPodSandbox(ctx, config.PodName, config.Namespace, config.UID, portMappings)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	// Prepare environment variables
	envs := map[string]string{
		"TYPE": config.ServerType,
		"EULA": "FALSE",
	}
	if config.EULA {
		envs["EULA"] = "TRUE"
	}
	
	// Add extra environment variables
	for k, v := range config.ExtraEnv {
		envs[k] = v
	}

	// Wait a moment for pod to be fully ready
	fmt.Printf("Waiting for pod to be ready...\n")
	time.Sleep(2 * time.Second)

	// Create container
	fmt.Printf("Creating container: %s\n", config.ContainerName)
	
	containerConfig := &runtime.ContainerConfig{
		Metadata: &runtime.ContainerMetadata{
			Name: config.ContainerName,
		},
		Image: &runtime.ImageSpec{
			Image: config.Image,
		},
		Envs: []*runtime.KeyValue{},
	}
	
	// Add environment variables
	for k, v := range envs {
		containerConfig.Envs = append(containerConfig.Envs, &runtime.KeyValue{
			Key:   k,
			Value: v,
		})
	}
	
	createReq := &runtime.CreateContainerRequest{
		PodSandboxId:  podID,
		Config:        containerConfig,
		SandboxConfig: podConfig,
	}
	
	createResp, err := c.RuntimeClient().CreateContainer(ctx, createReq)
	if err != nil {
		// Cleanup pod on failure
		_ = c.RemovePodSandbox(ctx, podID)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := createResp.ContainerId

	// Start container
	fmt.Printf("Starting container: %s\n", containerID)
	err = c.StartContainer(ctx, containerID)
	if err != nil {
		// Cleanup on failure
		_ = c.RemovePodSandbox(ctx, podID)
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✅ Minecraft server started successfully!\n")
	fmt.Printf("   Pod ID: %s\n", podID)
	fmt.Printf("   Container ID: %s\n", containerID)
	fmt.Printf("   Port: %d\n", config.HostPort)

	return &MinecraftServer{
		Client:      c,
		Config:      config,
		PodID:       podID,
		ContainerID: containerID,
	}, nil
}

// Stop stops the Minecraft server
func (m *MinecraftServer) Stop(ctx context.Context) error {
	fmt.Printf("Stopping Minecraft server (Pod: %s)\n", m.PodID)
	
	err := m.Client.StopPodSandbox(ctx, m.PodID)
	if err != nil {
		return fmt.Errorf("failed to stop pod: %w", err)
	}
	
	err = m.Client.RemovePodSandbox(ctx, m.PodID)
	if err != nil {
		return fmt.Errorf("failed to remove pod: %w", err)
	}
	
	fmt.Printf("✅ Minecraft server stopped successfully\n")
	return nil
}

// Status gets the current status of the Minecraft server
func (m *MinecraftServer) Status(ctx context.Context) (*runtime.ContainerStatus, error) {
	return m.Client.ContainerStatus(ctx, m.ContainerID)
}
