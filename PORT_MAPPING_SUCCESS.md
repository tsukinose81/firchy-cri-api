# ✅ ポートマッピング完全動作確認

## 実装完了

**CNI portmapプラグインを使用した自動ポートマッピングが完全に動作しています！**

### 動作確認結果

```bash
$ cd /home/dalai/firchy/cri-api/examples/minecraft
$ go run main.go 192.168.121.232:2375

Finding available port in range 1024-49151...
✅ Assigned host port: 18025
Port mapping: Container 25565 -> Host 18025
Creating pod sandbox: minecraft-pod
...
✅ Minecraft server started successfully!
   Container Port: 25565 (Minecraft default)
   Host Port: 18025

🎮 Connect to Minecraft server:
   Address: <VM_IP>:18025
   Example: 192.168.121.232:18025
```

### 接続確認

#### VM内から
```bash
$ sudo crictl ps | grep minecraft
1aa7d621c3cee   docker.io/itzg/minecraft-server:latest   Running   minecraft-server

# コンテナIPへ直接
$ timeout 3 bash -c "cat < /dev/null > /dev/tcp/10.88.0.39/25565"
✅ 成功

# ホストポート経由
$ timeout 3 bash -c "cat < /dev/null > /dev/tcp/localhost/18025"
✅ 成功
```

#### ホストマシンから
```bash
$ timeout 5 bash -c 'cat < /dev/null > /dev/tcp/192.168.121.232/18025'
✅ Minecraft accessible from host on port 18025!
```

### iptablesルール（自動生成）

```bash
$ sudo iptables -t nat -L -n -v | grep 18025
CNI-HOSTPORT-SETMARK  tcp  --  *  *  10.88.0.0/16   0.0.0.0/0  tcp dpt:18025
CNI-HOSTPORT-SETMARK  tcp  --  *  *  127.0.0.1      0.0.0.0/0  tcp dpt:18025
DNAT  tcp  --  *  *  0.0.0.0/0  0.0.0.0/0  tcp dpt:18025 to:10.88.0.39:25565
```

## 仕組み

### 1. ポート自動割り当て
```go
// port.go
func FindAvailablePort(minPort, maxPort int32) (int32, error) {
    // 1024-49151の範囲からランダムにポート選択
    port := minPort + rand.Int31n(portRange)
    return port, nil
}
```

### 2. PortMapping設定
```go
// minecraft.go
portMappings := []*runtime.PortMapping{
    {
        Protocol:      runtime.Protocol_TCP,  // 重要: 0 (TCP)
        ContainerPort: 25565,                 // コンテナ内は常に25565
        HostPort:      config.HostPort,       // 自動割り当てされたポート
    },
}
```

### 3. PodSandboxConfig
```go
podConfig := &runtime.PodSandboxConfig{
    Metadata: &runtime.PodSandboxMetadata{
        Name:      config.PodName,
        Namespace: config.Namespace,
        Uid:       config.UID,
    },
    Annotations: map[string]string{
        "io.containerd.cri.runtime-handler": "kata",
    },
    PortMappings: portMappings,  // ← ここでマッピング指定
    DnsConfig: &runtime.DNSConfig{
        Servers: []string{"8.8.8.8", "8.8.4.4"},
    },
}
```

### 4. CNI portmapプラグイン
- CRI APIの`port_mappings`を受け取る
- 自動的にiptables NATルールを生成
- PREROUTING、OUTPUT、POSTROUTINGチェーンに追加

## 重要なポイント

### ✅ 成功要因

1. **Protocol値**: `Protocol_TCP = 0` を使用（`Protocol_UDP = 1`ではない）
2. **同じpodConfig**: `RunPodSandbox`と`CreateContainer`で同じconfigを使用
3. **DNS設定**: Kata Containers VM内でのDNS解決を有効化
4. **待機時間**: Pod作成後2秒待機してから Container作成

### ❌ 以前の問題

1. `Protocol_TCP`を使っていたが、`RunPodSandbox()`が内部で別のconfigを作成していた
2. その結果、portMappingsが正しく適用されていなかった

## 使用方法

### デフォルト（ポート自動割り当て）
```go
config := criapi.DefaultMinecraftConfig()
// config.HostPort = 0  // 自動割り当て（デフォルト）

server, err := client.StartMinecraftServer(ctx, config)
// ログに割り当てられたポート番号が表示される
```

### 特定ポートを指定
```go
config := criapi.DefaultMinecraftConfig()
config.HostPort = 30000  // 特定のポートを指定

server, err := client.StartMinecraftServer(ctx, config)
```

## 接続方法

Minecraftクライアントで以下のアドレスに接続：
```
192.168.121.232:18025
```
（18025は自動割り当てされたポート番号）

## 追加設定不要

- ✅ iptablesルールは自動生成
- ✅ SSH接続不要
- ✅ 手動設定不要
- ✅ CNI portmapが全て処理

完全自動化されたポートマッピングが実現しました！
