# ポート転送の実装状況

## 現在の実装

### ✅ 完了
1. **ポート自動割り当て** - 1024-49151の範囲からランダムにポートを選択
2. **ログ出力** - 割り当てられたポートを明確に表示
3. **コンテナポート固定** - Minecraftは常に25565（標準ポート）

### ⚠️ 制限事項

**CNI Port Mappingの制限**

CRI APIの`port_mappings`を使用していますが、これだけではVM外部からのアクセスには不十分です。
以下の追加設定が必要：

1. **iptables NAT ルール** - VMホストでの設定
   ```bash
   sudo iptables -t nat -A PREROUTING -p tcp --dport <HOST_PORT> -j DNAT --to-destination <CONTAINER_IP>:25565
   sudo iptables -t nat -A OUTPUT -p tcp --dport <HOST_PORT> -j DNAT --to-destination <CONTAINER_IP>:25565
   sudo iptables -t nat -A POSTROUTING -p tcp -d <CONTAINER_IP> --dport 25565 -j MASQUERADE
   ```

2. **永続化** - VM再起動後も維持するには
   ```bash
   sudo apt-get install iptables-persistent
   sudo netfilter-persistent save
   ```

## 使用例

```go
// デフォルト設定（ポート自動割り当て）
config := criapi.DefaultMinecraftConfig()
// config.HostPort = 0  // 自動割り当て

server, err := client.StartMinecraftServer(ctx, config)

// 出力例:
// Finding available port in range 1024-49151...
// ✅ Assigned host port: 3170
// Port mapping: Container 25565 -> Host 3170
// ...
// 🎮 Connect to Minecraft server:
//    Address: <VM_IP>:3170
//    Example: 192.168.121.232:3170
```

## 今後の改善案

1. **自動iptables設定**
   - Goクライアントから SSH/API でVM内のiptablesを自動設定
   - またはAnsibleロールでサービス化

2. **ポート管理サービス**
   - 使用中のポートを追跡
   - ポート衝突を回避

3. **Kubernetesスタイルの Service**
   - NodePort/LoadBalancer風の実装
