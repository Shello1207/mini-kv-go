# mini-kv-go

一个基于 Go 实现的轻量级 KV 存储系统，目标是逐步演进为支持持久化、分片与复制的分布式 KV。

---

## ✨ Features

### Core (已完成)

* 单机 KV：支持 `PUT / GET / DELETE`
* 并发安全内存存储（`map + RWMutex`）
* HTTP API（基于标准库实现）

### Storage Engine

* **WAL（Write-Ahead Log）日志持久化**
* **Segmented WAL（多段日志）**

  * 基于文件大小自动轮转
  * 避免单文件无限增长
* **Crash Recovery（崩溃恢复）**

  * 启动时按顺序重放多个 WAL 文件
  * 支持跨文件数据恢复
* 删除操作通过 WAL 记录逻辑删除（类似 tombstone 机制），为后续 Bitcask 存储引擎设计预留

---

## 🧱 Architecture

### 写入流程（Write Path）

```text
Client → HTTP API → KV Service
                     ↓
                 Append WAL
                     ↓
                 MemTable (map)
```

### 启动恢复（Recovery）

```text
Scan WAL files → Sort → Replay → Rebuild MemTable
```

---

## 📦 API

| Method | Endpoint  | Description |
| ------ | --------- | ----------- |
| PUT    | /kv/{key} | 写入键值        |
| GET    | /kv/{key} | 读取键值        |
| DELETE | /kv/{key} | 删除键         |
| GET    | /health   | 健康检查        |

---

## 🚀 Quick Start

```bash
go run ./cmd/server
```

---

## 🧪 Example

```bash
curl -X PUT http://localhost:8080/kv/name -d "alice"
curl http://localhost:8080/kv/name
curl -X DELETE http://localhost:8080/kv/name
```

---

## 📁 Data Layout

```text
data/
  wal-1.log
  wal-2.log
  wal-3.log
```

* 日志采用 append-only 模式
* 超过阈值自动切分为多个 segment
* 启动时按顺序重放所有日志文件

---

## 📈 Current Status

* [x] In-memory KV store
* [x] WAL (Write-Ahead Log)
* [x] Segmented WAL with log rotation
* [x] Crash recovery across multiple log files
* [ ] Bitcask-style storage engine
* [ ] Compaction
* [ ] Consistent hashing sharding
* [ ] Leader-follower replication
* [ ] Docker Compose multi-node cluster

---

## 🛣 Roadmap

下一阶段计划实现：

* Bitcask 风格存储引擎（KeyDir + Data File）
* 数据文件 Compaction（合并与清理）
* 一致性哈希分片
* 主从复制（Leader-Follower）
* 多节点集群部署

---

## 💡 Design Notes

* 使用 append-only WAL 提升写入性能并简化恢复逻辑
* 通过 segmented log 降低单文件膨胀问题
* 当前采用 JSON 编码日志，便于调试，后续可优化为二进制格式
