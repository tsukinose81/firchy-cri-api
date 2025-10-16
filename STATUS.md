# firchy-cri-api 開発状況

## ✅ 完成した機能

### 1. CRI API クライアント実装
- containerd CRI API経由でPod/Container管理
- Kata Containers runtimeサポート
- TCP/Unix socket 接続対応
- **DNS設定サポート** - Pod作成時にDNSサーバーを指定可能

### 2. 動作確認済み
**環境:**
- containerd v1.7.28
- Kata Containers (Cloud Hypervisor)
- Remote TCP接続: 192.168.121.232:2375

**成功したテスト:**

#### Busyboxコンテナ
```bash
$ cd examples/busybox
$ go run main.go 192.168.121.232:2375

Container state: CONTAINER_RUNNING
```

#### Minecraftサーバー 🎉
```bash
$ cd examples/minecraft
$ go run main.go 192.168.121.232:2375

✅ Minecraft server started successfully!
   Port: 25565
Container Status:
  State: CONTAINER_RUNNING
```

**VM上での確認:**
```
$ sudo crictl ps
CONTAINER           IMAGE                                    CREATED             STATE               NAME
63b5da47bb3d5       docker.io/itzg/minecraft-server:latest   2 minutes ago       Running             minecraft

$ sudo crictl exec 63b5da47bb3d5 ps aux | grep java
minecra+     316 59.6 56.8 3261856 1158068 ?     Sl   07:52   0:53 java -Xmx1G -Xms1G -jar /data/paper-1.21.8-60.jar
```

**確認事項:**
- ✅ Pod作成成功（Kata runtime指定）
- ✅ DNS設定適用（8.8.8.8, 8.8.4.4）
- ✅ Container作成・起動成功  
- ✅ Cloud Hypervisor VM起動
- ✅ コンテナが `CONTAINER_RUNNING` 状態で動作
- ✅ **Minecraftサーバー（Paper 1.21.8）が正常に起動**

## 🔧 重要な修正

### DNS設定の追加
Kata Containers VM内でDNS解決を有効にするため、PodSandboxConfigに以下を追加：

```go
DnsConfig: &runtime.DNSConfig{
    Servers: []string{"8.8.8.8", "8.8.4.4"},
}
```

これにより、Minecraftサーバーが起動時に必要なファイル（Paper）をダウンロードできるようになりました。

## 📁 ファイル構成

```
/home/dalai/firchy/cri-api/
├── client.go          # CRI APIクライアント実装
├── minecraft.go       # Minecraftサーバー管理 ✅ 動作確認済み
├── README.md          # 基本ドキュメント
├── USAGE.md          # 使用ガイド
├── STATUS.md          # このファイル
├── examples/
│   ├── busybox/      # ✅ 動作確認済み - シンプルなコンテナ起動例
│   ├── list/         # Pod/Container一覧表示
│   └── minecraft/    # ✅ 動作確認済み - Minecraftサーバー起動
└── go.mod
```

## 🚀 使用方法

### Minecraftサーバー起動（完全動作）

```bash
cd /home/dalai/firchy/cri-api/examples/minecraft
go run main.go 192.168.121.232:2375
```

### 動作するサンプル（busybox）

```bash
cd /home/dalai/firchy/cri-api/examples/busybox
go run main.go 192.168.121.232:2375
```

### APIの使用例

```go
package main

import (
    "context"
    criapi "github.com/tsukinose81/firchy-cri-api"
    runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func main() {
    ctx := context.Background()
    client, _ := criapi.NewClient("192.168.121.232:2375")
    defer client.Close()

    // Podサンドボックス作成
    podConfig := &runtime.PodSandboxConfig{
        Metadata: &runtime.PodSandboxMetadata{
            Name:      "my-pod",
            Namespace: "default",
            Uid:       "my-uid-123",
        },
        Annotations: map[string]string{
            "io.containerd.cri.runtime-handler": "kata",
        },
    }
    
    podReq := &runtime.RunPodSandboxRequest{
        Config:         podConfig,
        RuntimeHandler: "kata",
    }
    
    podResp, _ := client.RuntimeClient().RunPodSandbox(ctx, podReq)
    podID := podResp.PodSandboxId

    // コンテナ作成
    containerConfig := &runtime.ContainerConfig{
        Metadata: &runtime.ContainerMetadata{
            Name: "my-container",
        },
        Image: &runtime.ImageSpec{
            Image: "docker.io/library/busybox:latest",
        },
        Command: []string{"/bin/sh", "-c", "sleep 3600"},
    }
    
    createReq := &runtime.CreateContainerRequest{
        PodSandboxId:  podID,
        Config:        containerConfig,
        SandboxConfig: podConfig,
    }
    
    createResp, _ := client.RuntimeClient().CreateContainer(ctx, createReq)
    containerID := createResp.ContainerId
    
    // コンテナ起動
    startReq := &runtime.StartContainerRequest{
        ContainerId: containerID,
    }
    client.RuntimeClient().StartContainer(ctx, startReq)
}
```

## 次のステップ

1. ✅ ~~Kata Containers のネットワーク/DNS設定を修正~~ **完了！**
2. ✅ ~~Minecraftサーバーの動作確認~~ **完了！**
3. より多くのサンプル追加（Nginx, Redis等）
4. エラーハンドリングの改善
5. ログ取得機能の実装（現在ログパスの問題あり）
6. ポート転送・ネットワーク設定の拡張
