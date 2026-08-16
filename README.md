# 个人运动打卡 API

一个使用 Gin 和 SQLite 构建的个人运动打卡服务。项目按 Go 企业分层结构组织，支持运动记录的增删改查、日期范围查询、按运动类型汇总，以及最近一周打卡查看。

## 技术栈

- Go 1.26
- Gin
- SQLite（modernc.org/sqlite 纯 Go 驱动）

## 目录结构

```text
cmd/server/main.go
internal/config
internal/model
internal/repository
internal/service
internal/handler
internal/router
internal/middleware
migrations
```

## 快速开始

```bash
go mod tidy
go test ./...
go run ./cmd/server
```

服务默认监听 `18009` 端口。可以通过环境变量覆盖：

```bash
PORT=18009 DB_PATH=fitness.db GIN_MODE=release go run ./cmd/server
```

## API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/workouts` | 新增打卡 |
| GET | `/api/v1/workouts` | 分页查询，支持 `sport_type`、`start_date`、`end_date` |
| GET | `/api/v1/workouts/:id` | 查询单条记录 |
| PUT | `/api/v1/workouts/:id` | 修改打卡 |
| DELETE | `/api/v1/workouts/:id` | 删除打卡 |
| GET | `/api/v1/workouts/summary/by-sport-type` | 按运动类型汇总时长和热量 |
| GET | `/api/v1/workouts/recent-week` | 查看最近一周打卡 |
| GET | `/healthz` | 健康检查 |

新增或修改请求体示例：

```json
{
  "sport_type": "running",
  "duration_minutes": 30,
  "workout_date": "2026-08-16",
  "note": "morning run"
}
```

## 热量规则

热量由 `internal/service` 层统一计算，规则为运动类型每分钟热量系数乘以运动时长。已配置的常见类型：

| 运动类型 | 每分钟热量 |
| --- | ---: |
| running | 10 |
| cycling | 8 |
| swimming | 9 |
| walking | 5 |
| yoga | 4 |
| hiking | 7 |
| strength | 6 |
| 其他 | 6 |

客户端无需传入 `calories`，服务端会根据 `sport_type` 和 `duration_minutes` 自动生成并覆盖该字段。

## 测试

```bash
go test ./...
```
