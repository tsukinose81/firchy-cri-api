# firchy-cri-api

GoでCRI APIを使用してKata ContainersでMinecraftサーバーを起動するためのライブラリです。

## 機能

- CRI (Container Runtime Interface) APIクライアント
- Kata Containers runtimeでのPod/Container管理
- Minecraftサーバーの簡単起動

## インストール

```bash
go get github.com/tsukinose81/firchy-cri-api
```

## 使い方

### 基本的な使用例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    criapi "github.com/tsukinose81/firchy-cri-api"
)

func main() {
    ctx := context.Background()

    // CRI APIクライアントを作成（Unix socket）
    client, err := criapi.NewClient("unix:///run/containerd/containerd.sock")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // Minecraftサーバーを起動
    config := criapi.DefaultMinecraftConfig()
    config.ServerType = "PAPER"
    config.HostPort = 25565

    server, err := client.StartMinecraftServer(ctx, config)
    if err != nil {
        log.Fatalf("Failed to start Minecraft server: %v", err)
    }

    // ステータス確認
    status, err := server.Status(ctx)
    if err != nil {
        log.Fatalf("Failed to get status: %v", err)
    }
    fmt.Printf("Container state: %v\n", status.State)

    // サーバーを実行し続ける
    time.Sleep(time.Hour)

    // 停止
    if err := server.Stop(ctx); err != nil {
        log.Fatalf("Failed to stop server: %v", err)
    }
}
```

### カスタム設定

```go
config := &criapi.MinecraftServerConfig{
    PodName:       "my-minecraft",
    Namespace:     "gaming",
    UID:           "mc-001",
    ContainerName: "minecraft",
    Image:         "docker.io/itzg/minecraft-server:latest",
    ServerType:    "PAPER",
    EULA:          true,
    ServerPort:    25565,
    HostPort:      25565,
    ExtraEnv: map[string]string{
        "MEMORY": "2G",
        "DIFFICULTY": "hard",
        "MAX_PLAYERS": "20",
    },
}

server, err := client.StartMinecraftServer(ctx, config)
```

### リモート接続

```go
// TCP経由で接続（containerdがTCPソケットで公開されている場合）
client, err := criapi.NewClient("192.168.121.232:2375")
```

### コマンドラインからの実行

```bash
# ローカルで実行（Unixソケット）
cd examples/minecraft
go run main.go unix:///run/containerd/containerd.sock

# リモートで実行（TCP）
go run main.go 192.168.121.232:2375

# Pod/Container一覧表示
cd examples/list
go run main.go 192.168.121.232:2375
```

## API

### Client

- `NewClient(endpoint string)` - CRI APIクライアントを作成
- `Close()` - 接続を閉じる
- `Version(ctx)` - ランタイムバージョンを取得
- `PullImage(ctx, image)` - イメージをプル
- `RunPodSandbox(ctx, name, namespace, uid, portMappings)` - Podを作成
- `CreateContainer(ctx, podID, name, image, command, envs)` - コンテナを作成
- `StartContainer(ctx, containerID)` - コンテナを起動
- `StopPodSandbox(ctx, podID)` - Podを停止
- `RemovePodSandbox(ctx, podID)` - Podを削除
- `ListPodSandbox(ctx)` - Pod一覧を取得
- `ListContainers(ctx)` - コンテナ一覧を取得
- `ContainerStatus(ctx, containerID)` - コンテナステータスを取得

### MinecraftServer

- `StartMinecraftServer(ctx, config)` - Minecraftサーバーを起動
- `Stop(ctx)` - サーバーを停止
- `Status(ctx)` - サーバーステータスを取得

### MinecraftServerConfig

- `DefaultMinecraftConfig()` - デフォルトの設定を取得
- `Validate()` - 設定が有効かどうかを検証

## 要件

- Go 1.21+
- containerd with CRI plugin
- Kata Containers runtime

## ライセンス

MIT
