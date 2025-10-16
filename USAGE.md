# firchy-cri-api 使用ガイド

## 検証済みの動作

リモートからGoクライアントでKata ContainersベースのMinecraftサーバーを起動できることを確認しました。

### 環境
- containerd: v1.7.28
- Kata Containers runtime (Cloud Hypervisor)
- TCP接続: 192.168.121.232:2375

### 実行結果

```bash
$ cd examples/minecraft
$ go run main.go 192.168.121.232:2375

Connecting to CRI API: 192.168.121.232:2375
Runtime: containerd v1.7.28

=== Starting Minecraft Server ===
Pulling image: docker.io/itzg/minecraft-server:latest
Creating pod sandbox: minecraft-pod
Creating container: minecraft-server
Starting container: xxx
✅ Minecraft server started successfully!
```

### 確認コマンド

```bash
# リモートVM上で確認
vagrant ssh

# Pod一覧
sudo crictl pods
# OUTPUT:
# POD ID              CREATED             STATE               NAME                NAMESPACE           ATTEMPT             RUNTIME
# 2c672eefbe203       40 minutes ago      Ready               minecraft-pod       default             0                   kata

# コンテナ一覧
sudo crictl ps
# OUTPUT:
# CONTAINER ID        IMAGE                                    CREATED             STATE               NAME                ATTEMPT             POD ID
# fa0284c57fdb9       docker.io/itzg/minecraft-server:latest   40 minutes ago      Running             mc                  0                   2c672eefbe203

# Cloud Hypervisorプロセス確認
ps aux | grep cloud-hypervisor | grep -v grep
# OUTPUT:
# root  21421  4.3 38.1 2385452 1529444 ?  Sl   06:30   1:46 /opt/kata/bin/cloud-hypervisor --api-socket /run/vc/vm/.../clh-api.sock

# Minecraftサーバーログ
sudo crictl logs fa0284c57fdb9 | tail -5
# OUTPUT:
# [06:31:12 INFO]: RCON running on 0.0.0.0:25575
# [06:31:12 INFO]: Running delayed init tasks
# [06:31:12 INFO]: Done (47.509s)! For help, type "help"
```

## Go API使用例

### 基本的なMinecraftサーバー起動

```go
package main

import (
    "context"
    "log"
    criapi "github.com/tsukinose81/firchy-cri-api"
)

func main() {
    ctx := context.Background()
    
    // リモート接続
    client, err := criapi.NewClient("192.168.121.232:2375")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // デフォルト設定で起動
    server, err := client.StartMinecraftServer(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // サーバー情報
    log.Printf("Pod ID: %s", server.PodID)
    log.Printf("Container ID: %s", server.ContainerID)
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
        "MEMORY":      "4G",
        "DIFFICULTY":  "hard",
        "MAX_PLAYERS": "20",
        "MOTD":        "Welcome to my server!",
        "MODE":        "survival",
    },
}

server, err := client.StartMinecraftServer(ctx, config)
```

### Pod/Container一覧取得

```go
// Pod一覧
pods, err := client.ListPodSandbox(ctx)
for _, pod := range pods {
    log.Printf("Pod: %s (State: %v)", pod.Metadata.Name, pod.State)
}

// Container一覧
containers, err := client.ListContainers(ctx)
for _, c := range containers {
    log.Printf("Container: %s (State: %v)", c.Metadata.Name, c.State)
}
```

### サーバー停止

```go
// 起動したサーバーを停止
err = server.Stop(ctx)
if err != nil {
    log.Fatal(err)
}
```

## コマンドライン実行

### Minecraftサーバー起動

```bash
cd examples/minecraft
go run main.go 192.168.121.232:2375
# Ctrl+C で停止
```

### リソース一覧表示

```bash
cd examples/list
go run main.go 192.168.121.232:2375
```

## トラブルシューティング

### 接続タイムアウト

```
failed to connect to ...: context deadline exceeded
```

→ containerdがTCPソケットでリッスンしているか確認
```bash
vagrant ssh -c "sudo ss -tlnp | grep 2375"
```

### コンテナ作成エラー

```
failed to create container: rpc error: code = Unavailable desc = error reading from server: EOF
```

→ Podが完全に起動するまで待つ（コードに2秒待機を追加済み）

### イメージPullエラー

VM内のネットワーク設定を確認してください。
