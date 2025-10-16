# 🎉 成功！Minecraftサーバーが起動しました

## 実行結果

**日時:** 2025-10-16 07:52 UTC

### Goクライアントからの起動

```bash
$ cd /home/dalai/firchy/cri-api/examples/minecraft
$ go run main.go 192.168.121.232:2375

Connecting to CRI API: 192.168.121.232:2375
Runtime: containerd v1.7.28

=== Starting Minecraft Server ===
Pulling image: docker.io/itzg/minecraft-server:latest
Creating pod sandbox: minecraft-pod
Waiting for pod to be ready...
Creating container: minecraft-server
Starting container: 63b5da47bb3d56662c21ec48ff71a2c69949cc849fe95fdab2d86f2f2f578622
✅ Minecraft server started successfully!
   Pod ID: 4f3155fb2f79553b04379d8309163faea68f7335e47c0e4eb6c581c0acb3a0d8
   Container ID: 63b5da47bb3d56662c21ec48ff71a2c69949cc849fe95fdab2d86f2f2f578622
   Port: 25565

Container Status:
  State: CONTAINER_RUNNING
```

### VM上での確認

```bash
vagrant ssh

# コンテナ確認
$ sudo crictl ps
CONTAINER           IMAGE                                    CREATED             STATE               NAME
63b5da47bb3d5       docker.io/itzg/minecraft-server:latest   2 minutes ago       Running             minecraft

# DNS設定確認
$ sudo crictl exec 63b5da47bb3d5 cat /etc/resolv.conf
nameserver 8.8.8.8
nameserver 8.8.4.4

# Java プロセス確認
$ sudo crictl exec 63b5da47bb3d5 ps aux | grep java
minecra+       2  0.0  0.2 1229724 5480 ?        Sl   07:52   0:00 mc-server-runner --stop-duration 60s java -Xmx1G -Xms1G -jar /data/paper-1.21.8-60.jar
minecra+     316 59.6 56.8 3261856 1158068 ?     Sl   07:52   0:53 java -Xmx1G -Xms1G -jar /data/paper-1.21.8-60.jar

# Cloud Hypervisor確認
$ ps aux | grep cloud-hypervisor | grep -v grep
root       31895  5.1 38.3 2385452 1557332 ?     Sl   07:52   0:28 /opt/kata/bin/cloud-hypervisor --api-socket /run/vc/vm/.../clh-api.sock
```

## 成功の鍵

### DNS設定の追加

Kata Containers VM内でのDNS解決を有効にするため、PodSandboxConfigにDNS設定を追加しました：

```go
// minecraft.go
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
        Servers: []string{"8.8.8.8", "8.8.4.4"},  // ← これが重要！
    },
}
```

### トラブルシューティングの過程

1. **問題発見**: busyboxコンテナから `curl ip-api.com` を実行
   - 結果: `Could not resolve host` エラー
   
2. **原因特定**: コンテナ内の `/etc/resolv.conf` を確認
   - `nameserver 127.0.0.53` (systemd-resolved) がKata VM内で機能していなかった
   
3. **解決策**: PodSandboxConfigでDNSサーバーを明示的に指定
   - Google Public DNS (8.8.8.8, 8.8.4.4) を設定

4. **検証**: busyboxコンテナで再テスト
   - `nslookup google.com` 成功
   - `wget http://google.com` 成功
   
5. **最終確認**: Minecraftサーバー起動
   - Paper 1.21.8のダウンロード成功
   - サーバー起動成功 ✅

## 技術スタック

- **Go**: 1.21+
- **CRI API**: k8s.io/cri-api v0.34.1
- **gRPC**: google.golang.org/grpc v1.76.0
- **containerd**: v1.7.28
- **Kata Containers**: Cloud Hypervisor
- **Minecraft**: Paper 1.21.8 (via itzg/minecraft-server:latest)

## 成果物

✅ リモートからGoでKata ContainersのMinecraftサーバーを起動できるCRI APIクライアント
✅ TCP接続による遠隔操作
✅ DNS設定サポート
✅ Pod/Container管理機能
✅ 完全に動作するサンプルコード

すべて `/home/dalai/firchy/cri-api` に配置されています！
