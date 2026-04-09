# mini-kv-go

一个基于 Go 实现的轻量级 KV 存储系统，目标是逐步演进为支持持久化、分片与复制的分布式 KV。

## Features (Week 1)
- 单机 KV：PUT / GET / DELETE
- 并发安全内存存储（map + RWMutex）
- WAL（Write-Ahead Log）日志持久化
- 启动时日志重放恢复
- HTTP API

## API
PUT /kv/{key}
GET /kv/{key}
DELETE /kv/{key}
GET /health

## Quick Start
go run ./cmd/server

## Example
curl -X PUT http://localhost:8080/kv/name -d "alice"
curl http://localhost:8080/kv/name
curl -X DELETE http://localhost:8080/kv/name

## Roadmap
- [x] In-memory KV
- [x] WAL + Recovery
- [ ] Bitcask-style storage engine
- [ ] Compaction
- [ ] Consistent hashing sharding
- [ ] Leader-follower replication
- [ ] Docker Compose multi-node cluster
