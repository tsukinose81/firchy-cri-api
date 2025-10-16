package criapi

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// Client is a CRI API client for managing containers with Kata runtime
type Client struct {
	conn            *grpc.ClientConn
	runtimeClient   runtime.RuntimeServiceClient
	imageClient     runtime.ImageServiceClient
	endpoint        string
}

// RuntimeClient returns the underlying runtime service client
func (c *Client) RuntimeClient() runtime.RuntimeServiceClient {
	return c.runtimeClient
}

// ImageClient returns the underlying image service client
func (c *Client) ImageClient() runtime.ImageServiceClient {
	return c.imageClient
}

// NewClient creates a new CRI API client
func NewClient(endpoint string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}

	return &Client{
		conn:          conn,
		runtimeClient: runtime.NewRuntimeServiceClient(conn),
		imageClient:   runtime.NewImageServiceClient(conn),
		endpoint:      endpoint,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Version returns the runtime version
func (c *Client) Version(ctx context.Context) (*runtime.VersionResponse, error) {
	req := &runtime.VersionRequest{
		Version: "v1",
	}
	return c.runtimeClient.Version(ctx, req)
}

// PullImage pulls a container image
func (c *Client) PullImage(ctx context.Context, image string) (string, error) {
	req := &runtime.PullImageRequest{
		Image: &runtime.ImageSpec{
			Image: image,
		},
	}
	
	resp, err := c.imageClient.PullImage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	
	return resp.ImageRef, nil
}

// RunPodSandbox creates and starts a pod sandbox with Kata runtime
func (c *Client) RunPodSandbox(ctx context.Context, name, namespace, uid string, portMappings []*runtime.PortMapping) (string, error) {
	config := &runtime.PodSandboxConfig{
		Metadata: &runtime.PodSandboxMetadata{
			Name:      name,
			Namespace: namespace,
			Uid:       uid,
		},
		Annotations: map[string]string{
			"io.containerd.cri.runtime-handler": "kata",
		},
		PortMappings: portMappings,
	}

	req := &runtime.RunPodSandboxRequest{
		Config:         config,
		RuntimeHandler: "kata",
	}

	resp, err := c.runtimeClient.RunPodSandbox(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to run pod sandbox: %w", err)
	}

	return resp.PodSandboxId, nil
}

// CreateContainer creates a container in a pod sandbox
func (c *Client) CreateContainer(ctx context.Context, podID, containerName, image string, command []string, envs map[string]string, podConfig *runtime.PodSandboxConfig) (string, error) {
	var envVars []*runtime.KeyValue
	for k, v := range envs {
		envVars = append(envVars, &runtime.KeyValue{
			Key:   k,
			Value: v,
		})
	}

	config := &runtime.ContainerConfig{
		Metadata: &runtime.ContainerMetadata{
			Name: containerName,
		},
		Image: &runtime.ImageSpec{
			Image: image,
		},
		Command: command,
		Envs:    envVars,
		Stdin:   false,
		StdinOnce: false,
		Tty:     false,
	}

	req := &runtime.CreateContainerRequest{
		PodSandboxId:  podID,
		Config:        config,
		SandboxConfig: podConfig,
	}

	resp, err := c.runtimeClient.CreateContainer(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ContainerId, nil
}

// StartContainer starts a container
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	req := &runtime.StartContainerRequest{
		ContainerId: containerID,
	}

	_, err := c.runtimeClient.StartContainer(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// StopPodSandbox stops a pod sandbox
func (c *Client) StopPodSandbox(ctx context.Context, podID string) error {
	req := &runtime.StopPodSandboxRequest{
		PodSandboxId: podID,
	}

	_, err := c.runtimeClient.StopPodSandbox(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stop pod sandbox: %w", err)
	}

	return nil
}

// RemovePodSandbox removes a pod sandbox
func (c *Client) RemovePodSandbox(ctx context.Context, podID string) error {
	req := &runtime.RemovePodSandboxRequest{
		PodSandboxId: podID,
	}

	_, err := c.runtimeClient.RemovePodSandbox(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to remove pod sandbox: %w", err)
	}

	return nil
}

// ListPodSandbox lists pod sandboxes
func (c *Client) ListPodSandbox(ctx context.Context) ([]*runtime.PodSandbox, error) {
	req := &runtime.ListPodSandboxRequest{}

	resp, err := c.runtimeClient.ListPodSandbox(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod sandboxes: %w", err)
	}

	return resp.Items, nil
}

// ListContainers lists containers
func (c *Client) ListContainers(ctx context.Context) ([]*runtime.Container, error) {
	req := &runtime.ListContainersRequest{}

	resp, err := c.runtimeClient.ListContainers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return resp.Containers, nil
}

// ContainerStatus gets container status
func (c *Client) ContainerStatus(ctx context.Context, containerID string) (*runtime.ContainerStatus, error) {
	req := &runtime.ContainerStatusRequest{
		ContainerId: containerID,
		Verbose:     true,
	}

	resp, err := c.runtimeClient.ContainerStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	return resp.Status, nil
}
