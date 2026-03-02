# telegram-bot

一个支持多业务场景的 Telegram Bot 框架，内置高性能 KV 存储系统。

## 特性

- 🤖 支持多 Bot 实例管理
- 💾 基于 WAL（Write-Ahead Logging）的持久化 KV 存储
- 🔧 灵活的配置系统
- 🚀 高性能：批量同步、自动压缩
- 🏢 多业务场景支持：每个业务独立存储文件

## KV 存储系统

### 设计理念

**一个文件一个 KVStore 实例，应用启动时统一加载，满足多业务场景需求。**

- **按业务隔离**：不同业务（用户数据、会话、缓存等）使用独立的 KVStore 实例
- **独立文件**：每个 KVStore 实例对应一个独立的存储文件
- **启动加载**：应用启动时统一初始化所有 KVStore 实例
- **无锁设计**：KVStoreManager 使用全局 map，启动后只读取，无并发写入场景，无需锁保护
- **无锁竞争**：每个 KVStore 实例独立管理，避免跨实例的锁竞争

### 存储架构

```
┌─────────────────────────────────────┐
│      KVStoreManager (管理器)        │
├─────────────────────────────────────┤
│  - user_topic → FileKVStore         │
│  - sessions   → FileKVStore         │
│  - users      → FileKVStore         │
│  - cache      → FileKVStore         │
└─────────────────────────────────────┘
         ↓              ↓
   user_topic.db   sessions.db  ...
```

### WAL 实现细节

- **数据格式**：`crc32 | op(put/del) | keyLen | valLen | key | value`
- **启动恢复**：顺序扫描日志文件，重建内存索引
- **写入策略**：追加写（append-only），支持强一致或批量同步
- **自动压缩**：定期清理已删除数据，节省磁盘空间

### 配置示例

```yaml
kvstores:
  - name: "user_topic"         # 业务名称（唯一标识）
    path: "./data/user_topic.db"  # 存储文件路径
    sync_on_write: false       # 不立即同步（提高性能）
    sync_threshold: 100        # 100次操作后触发同步
    compact_delete_count: 1000 # 1000次删除后触发压缩
    compact_cooldown: 10       # 压缩冷却时间（秒）
    sync_cooldown: 1           # 同步冷却时间（秒）
  
  - name: "sessions"
    path: "./data/sessions.db"
    sync_on_write: false
    sync_threshold: 100
    compact_delete_count: 1000
    compact_cooldown: 10
    sync_cooldown: 1
```

### 使用示例

```go
// 1. 应用启动时初始化（main.go）
cfg, _ := LoadConfig("config.yaml")
kvManager, _ := cfg.InitKVStoreManager()
defer kvManager.Close()

// 2. 创建存储工厂
storageFactory := NewKVStoreFactory(kvManager)

// 3. 在业务代码中使用
store, _ := storageFactory.GetBusinessStorage(BusinessUserTopic)

// 写入数据
store.Set("user_123", []byte(`{"name": "Alice"}`))

// 读取数据
data, found, _ := store.Get("user_123")
if found {
    fmt.Println(string(data))
}

// 删除数据
store.Delete("user_123")

// 遍历数据
store.ForEach(func(key string, value []byte) error {
    fmt.Printf("%s: %s\n", key, string(value))
    return nil
})
```

### 业务存储类型

框架预定义了以下业务存储：

- `user_topic`：用户话题映射（Bot使用）
- `sessions`：会话数据
- `users`：用户信息
- `cache`：临时缓存

可根据需要在 `config.yaml` 中添加更多业务存储。

## 配置

### config.yaml 结构

```yaml
proxy:
  enabled: true
  url: "socks5://127.0.0.1:7890"

bot:
  - token: "YOUR_BOT_TOKEN_1"
  - token: "YOUR_BOT_TOKEN_2"

customer:
  session_limit: 5000
  groups:
    - chat_id: -1003563520720

kvstores:
  - name: "user_topic"
    path: "./data/user_topic.db"
    # ... 其他配置
```

## 运行

```bash
# 构建
go build

# 运行
./telegram-bot
```

## 技术栈

- Go 1.21+
- gopkg.in/telebot.v4（Telegram Bot API）
- gopkg.in/yaml.v3（配置解析）
- 自研 WAL KV 存储引擎

## 性能特点

- **批量同步**：避免频繁 fsync，提高写入性能
- **内存索引**：快速定位数据，O(1) 查找
- **自动压缩**：后台清理，无需手动维护
- **并发安全**：内置读写锁，支持多 goroutine 访问
- **崩溃恢复**：自动截断损坏数据，保证一致性

## 许可

MIT License
