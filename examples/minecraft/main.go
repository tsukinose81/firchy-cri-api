package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	criapi "github.com/tsukinose81/firchy-cri-api"
)

func main() {
	ctx := context.Background()

	// CRI APIエンドポイント（デフォルトはUnix socket）
	endpoint := "unix:///run/containerd/containerd.sock"
	if len(os.Args) > 1 {
		endpoint = os.Args[1]
	}

	fmt.Printf("Connecting to CRI API: %s\n", endpoint)

	// クライアント作成
	client, err := criapi.NewClient(endpoint)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// バージョン確認
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}
	fmt.Printf("Runtime: %s %s\n", version.RuntimeName, version.RuntimeVersion)

	// Minecraftサーバー設定
	config := criapi.DefaultMinecraftConfig()
	config.ServerType = "PAPER"
	// HostPort = 0 for auto-assignment from 1024-49151
	config.HostPort = 0
	config.ExtraEnv = map[string]string{
		"MEMORY":      "2G",
		"DIFFICULTY":  "normal",
		"MAX_PLAYERS": "20",
		"MOTD":        "Kata Containers Minecraft Server",
	}

	// サーバー起動
	fmt.Println("\n=== Starting Minecraft Server ===")
	server, err := client.StartMinecraftServer(ctx, config)
	if err != nil {
		log.Fatalf("Failed to start Minecraft server: %v", err)
	}

	// ステータス確認
	time.Sleep(3 * time.Second)
	status, err := server.Status(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get status: %v", err)
	} else {
		fmt.Printf("\nContainer Status:\n")
		fmt.Printf("  State: %v\n", status.State)
		fmt.Printf("  Started At: %v\n", time.Unix(0, status.StartedAt))
		fmt.Printf("  Image: %s\n", status.Image.Image)
	}

	fmt.Printf("\n=== Minecraft Server is Running ===\n")
	fmt.Printf("Connect to: localhost:%d\n", config.HostPort)
	fmt.Printf("Press Ctrl+C to stop the server\n\n")

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\nReceived shutdown signal...")

	// サーバー停止
	fmt.Println("=== Stopping Minecraft Server ===")
	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	fmt.Println("Done!")
}
