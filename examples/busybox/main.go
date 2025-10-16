package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	criapi "github.com/tsukinose81/firchy-cri-api"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func main() {
	ctx := context.Background()

	endpoint := "192.168.121.232:2375"
	if len(os.Args) > 1 {
		endpoint = os.Args[1]
	}

	fmt.Printf("Connecting to: %s\n", endpoint)
	client, err := criapi.NewClient(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create pod
	fmt.Println("Creating pod...")
	podConfig := &runtime.PodSandboxConfig{
		Metadata: &runtime.PodSandboxMetadata{
			Name:      "test-pod-go",
			Namespace: "default",
			Uid:       "test-go-123",
		},
		Annotations: map[string]string{
			"io.containerd.cri.runtime-handler": "kata",
		},
		DnsConfig: &runtime.DNSConfig{
			Servers: []string{"8.8.8.8", "8.8.4.4"},
		},
	}

	podReq := &runtime.RunPodSandboxRequest{
		Config:         podConfig,
		RuntimeHandler: "kata",
	}

	podResp, err := client.RuntimeClient().RunPodSandbox(ctx, podReq)
	if err != nil {
		log.Fatalf("Failed to create pod: %v", err)
	}
	podID := podResp.PodSandboxId
	fmt.Printf("Pod ID: %s\n", podID)

	// Wait for pod to be ready
	time.Sleep(2 * time.Second)

	// Create container
	fmt.Println("Creating container...")
	containerConfig := &runtime.ContainerConfig{
		Metadata: &runtime.ContainerMetadata{
			Name: "busybox-test",
		},
		Image: &runtime.ImageSpec{
			Image: "docker.io/library/busybox:latest",
		},
		Command: []string{
			"/bin/sh",
			"-c",
			"echo 'Running in Kata from Go!' && sleep 60",
		},
	}

	createReq := &runtime.CreateContainerRequest{
		PodSandboxId:  podID,
		Config:        containerConfig,
		SandboxConfig: podConfig,
	}

	createResp, err := client.RuntimeClient().CreateContainer(ctx, createReq)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	containerID := createResp.ContainerId
	fmt.Printf("Container ID: %s\n", containerID)

	// Start container
	fmt.Println("Starting container...")
	startReq := &runtime.StartContainerRequest{
		ContainerId: containerID,
	}

	_, err = client.RuntimeClient().StartContainer(ctx, startReq)
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	fmt.Println("✅ Container started successfully!")

	// Wait and check status
	time.Sleep(5 * time.Second)

	statusReq := &runtime.ContainerStatusRequest{
		ContainerId: containerID,
	}
	statusResp, err := client.RuntimeClient().ContainerStatus(ctx, statusReq)
	if err != nil {
		log.Printf("Warning: failed to get status: %v", err)
	} else {
		fmt.Printf("Container state: %v\n", statusResp.Status.State)
	}

	fmt.Println("\nPress Ctrl+C to stop...")
	time.Sleep(120 * time.Second)

	// Cleanup
	fmt.Println("\nCleaning up...")
	client.StopPodSandbox(ctx, podID)
	client.RemovePodSandbox(ctx, podID)
	fmt.Println("Done!")
}
