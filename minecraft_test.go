package criapi

import (
	"testing"
)

func TestMinecraftServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *MinecraftServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			config:  DefaultMinecraftConfig(),
			wantErr: false,
		},
		{
			name: "empty pod name",
			config: &MinecraftServerConfig{
				PodName:       "",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "pod name cannot be empty",
		},
		{
			name: "empty namespace",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "namespace cannot be empty",
		},
		{
			name: "empty UID",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "UID cannot be empty",
		},
		{
			name: "empty container name",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "container name cannot be empty",
		},
		{
			name: "empty image",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "",
				ServerPort:    25565,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "image cannot be empty",
		},
		{
			name: "invalid server port - zero",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    0,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "server port must be between 1 and 65535",
		},
		{
			name: "invalid server port - too high",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    70000,
				HostPort:      25565,
			},
			wantErr: true,
			errMsg:  "server port must be between 1 and 65535",
		},
		{
			name: "invalid host port - zero",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      0,
			},
			wantErr: true,
			errMsg:  "host port must be between 1 and 65535",
		},
		{
			name: "invalid host port - too high",
			config: &MinecraftServerConfig{
				PodName:       "minecraft-pod",
				Namespace:     "default",
				UID:           "test-123",
				ContainerName: "minecraft",
				Image:         "docker.io/itzg/minecraft-server:latest",
				ServerPort:    25565,
				HostPort:      70000,
			},
			wantErr: true,
			errMsg:  "host port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				// Check if error message contains the expected substring
				if tt.errMsg != "" {
					errStr := err.Error()
					contains := false
					// For port errors, just check if it starts with the expected message
					if len(errStr) >= len(tt.errMsg) && errStr[:len(tt.errMsg)] == tt.errMsg {
						contains = true
					} else if errStr == tt.errMsg {
						contains = true
					}
					if !contains {
						t.Errorf("Validate() error = %q, want error message starting with %q", errStr, tt.errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDefaultMinecraftConfig(t *testing.T) {
	config := DefaultMinecraftConfig()
	
	if config == nil {
		t.Fatal("DefaultMinecraftConfig() returned nil")
	}
	
	// Verify default config is valid
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultMinecraftConfig() returned invalid config: %v", err)
	}
	
	// Verify default values
	if config.PodName == "" {
		t.Error("DefaultMinecraftConfig() PodName is empty")
	}
	if config.Namespace == "" {
		t.Error("DefaultMinecraftConfig() Namespace is empty")
	}
	if config.UID == "" {
		t.Error("DefaultMinecraftConfig() UID is empty")
	}
	if config.ContainerName == "" {
		t.Error("DefaultMinecraftConfig() ContainerName is empty")
	}
	if config.Image == "" {
		t.Error("DefaultMinecraftConfig() Image is empty")
	}
	if config.ServerPort != 25565 {
		t.Errorf("DefaultMinecraftConfig() ServerPort = %d, want 25565", config.ServerPort)
	}
	if config.HostPort != 25565 {
		t.Errorf("DefaultMinecraftConfig() HostPort = %d, want 25565", config.HostPort)
	}
	if config.ExtraEnv == nil {
		t.Error("DefaultMinecraftConfig() ExtraEnv is nil")
	}
}
