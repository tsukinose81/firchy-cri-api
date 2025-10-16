package main

import (
	"context"
	"fmt"
	"log"
	"os"

	criapi "github.com/tsukinose81/firchy-cri-api"
)

func main() {
	ctx := context.Background()

	endpoint := "unix:///run/containerd/containerd.sock"
	if len(os.Args) > 1 {
		endpoint = os.Args[1]
	}

	client, err := criapi.NewClient(endpoint)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// List all pods
	fmt.Println("=== Pod Sandboxes ===")
	pods, err := client.ListPodSandbox(ctx)
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	for _, pod := range pods {
		fmt.Printf("Pod: %s\n", pod.Metadata.Name)
		fmt.Printf("  ID: %s\n", pod.Id)
		fmt.Printf("  State: %v\n", pod.State)
		fmt.Printf("  Namespace: %s\n", pod.Metadata.Namespace)
		fmt.Println()
	}

	// List all containers
	fmt.Println("=== Containers ===")
	containers, err := client.ListContainers(ctx)
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	for _, container := range containers {
		fmt.Printf("Container: %s\n", container.Metadata.Name)
		fmt.Printf("  ID: %s\n", container.Id)
		fmt.Printf("  State: %v\n", container.State)
		fmt.Printf("  Image: %s\n", container.Image.Image)
		fmt.Printf("  Pod ID: %s\n", container.PodSandboxId)
		fmt.Println()
	}
}
