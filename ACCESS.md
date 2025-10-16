# Minecraftサーバーへのアクセス方法

## 🎮 接続アドレス

### ホストマシンから
```
アドレス: 192.168.121.232:25565
```

Minecraftクライアントで、マルチプレイヤー → サーバーを追加 → `192.168.121.232:25565`

### VM内部から
```bash
# 方法1: コンテナIPに直接接続
10.88.0.29:25565

# 方法2: localhost経由（iptables NAT）
localhost:25565
```

## ネットワーク構成

```
[Minecraft Client on Host]
        ↓
   192.168.121.232:25565 (Vagrant VM)
        ↓
   iptables DNAT
        ↓
   10.88.0.29:25565 (Kata Container)
        ↓
   Minecraft Server (Paper 1.21.8)
        ↓
   Cloud Hypervisor VM
```

## 接続確認

### ホストから
```bash
# ポートが開いているか確認
nc -zv 192.168.121.232 25565
# または
telnet 192.168.121.232 25565
```

### VM内から
```bash
vagrant ssh

# コンテナIPへ直接
nc -zv 10.88.0.29 25565

# localhostへ（iptables経由）
nc -zv localhost 25565
```

## サーバー情報

- **サーバータイプ:** Paper 1.21.8
- **実行環境:** Kata Containers (Cloud Hypervisor)
- **メモリ:** 1GB (デフォルト)
- **難易度:** normal
- **最大プレイヤー数:** 20

## トラブルシューティング

### 接続できない場合

1. **Minecraftサーバーが起動しているか確認:**
   ```bash
   vagrant ssh -c "sudo crictl ps | grep minecraft"
   ```

2. **iptables ルールが設定されているか確認:**
   ```bash
   vagrant ssh -c "sudo iptables -t nat -L -n -v | grep 25565"
   ```

3. **コンテナが実際にリッスンしているか確認:**
   ```bash
   vagrant ssh -c "sudo crictl exec 63b5da47bb3d5 netstat -tlnp | grep 25565"
   ```

### ファイアウォールの問題

ホストマシンにファイアウォールがある場合、192.168.121.232への接続を許可する必要があります。

## iptables ルールの永続化

現在のiptablesルールは再起動後に消えます。永続化するには:

```bash
vagrant ssh
sudo apt-get install iptables-persistent
sudo netfilter-persistent save
```

または、Ansibleロールに追加して自動設定します。
