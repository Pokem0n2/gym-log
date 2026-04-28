# gym-log

超轻量健身数据管理服务器

## 技术栈

- **Go 1.23** + **Gin**
- **SQLite** (纯 Go 实现，无 CGO 依赖)
- **Docker** 多阶段构建，最终镜像 `< 20MB`

## 快速开始

```bash
# 本地运行
go mod tidy
go run .

# Docker
docker-compose up -d
```

## API

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/exercises` | 动作列表 |
| POST | `/api/v1/exercises` | 新建动作 |
| DELETE | `/api/v1/exercises/:id` | 删除动作 |
| GET | `/api/v1/workouts` | 训练记录列表 |
| GET | `/api/v1/workouts/:id` | 训练详情（含组记录） |
| POST | `/api/v1/workouts` | 新建训练 |
| DELETE | `/api/v1/workouts/:id` | 删除训练 |
| POST | `/api/v1/workouts/:workout_id/sets` | 添加组 |
| DELETE | `/api/v1/sets/:id` | 删除组 |
| GET | `/api/v1/stats/exercise/:exercise_id` | 动作历史记录 |
| GET | `/api/v1/stats/volume?start=YYYY-MM-DD&end=YYYY-MM-DD` | 日期范围训练量统计 |
