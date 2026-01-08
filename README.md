# Docs Hub

API 文档聚合服务（文档中心），用于统一管理和展示多个微服务的 Swagger 文档。

## 部署方式

| 方式 | URL | 说明 |
|------|-----|------|
| **GitHub Pages** | https://YOUR_ORG.github.io/docs-hub/ | 静态预览，无需服务器 |
| **本地开发** | http://localhost:9000 | 实时拉取服务文档 |
| **Docker** | http://localhost:9000 | 容器化部署 |

## 功能特性

- 自动聚合多个微服务的 Swagger 文档
- 统一的 Swagger UI 展示界面
- 服务健康状态监控
- 定时自动刷新文档
- 支持手动刷新文档

## 工作机制

docs-hub 通过 **后台拉取 + 内存缓存** 的方式聚合各微服务的 API 文档：

```
┌───────────────────────────────────────────────────────────────┐
│                      docs-hub (9000)                          │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  DocService (后台拉取 + 缓存)                              │   │
│  │  1. 启动时从各服务拉取 swagger.json                       │   │
│  │  2. 定时刷新（refresh_interval: 1m/5m）                    │   │
│  │  3. 缓存在内存中                                         │   │
│  └────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  GET /api/docs/order-service/swagger.json → 返回缓存的内容    │
└───────────────────────────────────────────────────────────────┘
                               │
       ┌────────────────────┴────────────────────┐
       ▼                                            ▼
┌───────────────────┐               ┌────────────────────┐
│ user-service:8001  │               │ order-service:8002  │
│ /docs/swagger/     │               │ /docs/swagger/      │
│   swagger.json     │               │   swagger.json      │
└───────────────────┘               └────────────────────┘
```

### 服务发现配置

通过配置文件指定各微服务的地址：

**本地开发 (local.yaml)** - 直连本地端口：
```yaml
services:
  - name: "order-service"
    base_url: "http://localhost:8002"           # 本地端口
    doc_path: "/docs/swagger/swagger.json"      # 文档路径
```

**Docker Compose (dev.yaml)** - 使用服务名：
```yaml
services:
  - name: "order-service"
    base_url: "http://order-service:8002"       # Docker 网络服务名
    doc_path: "/docs/swagger/swagger.json"
```

### 文档拉取流程

1. **启动时**：DocService 从各服务拉取文档并缓存
2. **定时刷新**：根据 `refresh_interval` 配置定期更新
3. **手动刷新**：调用 `POST /api/refresh` 立即更新
4. **请求响应**：`GET /api/docs/:service/swagger.json` 返回缓存内容

## 快速开始

### Docker Compose 部署（推荐）

```bash
# 启动文档聚合服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### Docker Compose 开发模式

```bash
# 启动开发环境（连接本地微服务）
docker-compose -f docker-compose.dev.yml up -d
```

### 本地开发

```bash
# 安装依赖
go mod download

# 运行服务（连接本地微服务）
make run-local

# 访问文档门户
open http://localhost:9000
```

### Docker 运行

```bash
# 构建镜像
docker build -t docs-hub:latest .

# 运行容器
docker run -p 9000:9000 docs-hub:latest
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/catalog` | GET | 获取服务目录 |
| `/api/docs/:service/swagger.json` | GET | 获取服务 Swagger JSON |
| `/api/services/:service` | GET | 获取服务详情 |
| `/api/refresh` | POST | 刷新所有文档 |
| `/api/refresh/:service` | POST | 刷新单个服务文档 |

## 目录结构

```
docs-hub/
├── cmd/
│   └── main.go              # 应用入口
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 处理器
│   └── service/             # 业务逻辑
├── web/
│   └── templates/           # HTML 模板
├── configs/                 # 配置文件
├── Dockerfile               # Docker 配置
├── Makefile                 # 构建脚本
└── go.mod                   # Go 模块文件
```

## 配置说明

| 环境 | 配置文件 | 说明 |
|------|---------|------|
| 本地开发 | local.yaml | 连接本地运行的微服务 |
| 远程开发 | dev.yaml | Docker Compose 环境 |
| 生产环境 | prod.yaml | Kubernetes 部署 |
| **GitHub Pages** | services.json | 配置各服务的 Swagger URL |

### GitHub Pages 服务配置

GitHub Pages 部署使用 **环境变量** 动态生成配置，避免污染 main 分支：

#### 配置步骤

1. 创建 Environment：**Settings → Environments → New environment** → 输入 `github-pages`

2. 添加 Variables（可选，用于自定义配置）：

| Variable | 说明 | 示例 |
|----------|------|------|
| `SERVICES_CONFIG` | 自定义服务配置（可选） | 见下方格式 |

#### 默认行为

如果未配置环境变量，CI 会自动使用 `github.repository_owner` 推断 URL：

```
https://{owner}.github.io/user-service/swagger.json
https://{owner}.github.io/order-service/swagger.json
```

#### 自定义配置格式

如需自定义服务列表，设置 `SERVICES_CONFIG` 变量（格式：`name|title|description|swaggerUrl`）：

```
user-service|用户服务|用户管理|https://myorg.github.io/user-service/swagger.json
order-service|订单服务|订单管理|https://myorg.github.io/order-service/swagger.json
```

> 本地模板文件 `configs/services.json` 仅作为参考，CI 会从环境变量动态生成。

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| CONFIG_PATH | 配置文件路径 | configs/dev.yaml |
| PORT | 服务端口 | 9000 |
| GIN_MODE | Gin 运行模式 | debug |
