
</think>

# Tavily 代理池 & 管理面板

简体中文 | [English](./README_EN.md)

> 以下链接默认使用仓库路径 `yunshenwuchuxun/TavilyProxy`。如果你最终发布时使用了不同仓库名，请一并替换下方链接。

[![Build](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/build-multi-platform.yml/badge.svg)](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/build-multi-platform.yml)
[![Docker Publish](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/docker-publish.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)](https://nodejs.org/)
[![Render Blueprint](https://img.shields.io/badge/Render-Blueprint-46E3B7?logo=render&logoColor=black)](./render.yaml)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/yunshenwuchuxun/TavilyProxy)

一个透明的 Tavily API 反向代理：将多个 Tavily API Key（额度/credits）汇聚在一个 **Master Key** 之后，并提供内置 Web UI 用于管理 Key、用量与请求日志。

> **快速区分**
> - **管理面板登录**：首次登录默认是 `admin / admin`；如果你在首次启动前设置了 `ADMIN_USERNAME` / `ADMIN_PASSWORD`，或数据库里已经保存过管理员凭据，则以那些值为准。
> - **Master Key**：用于 `/search`、`/extract`、`/crawl`、`/mcp` 等代理调用，**不用于管理面板登录**。
> - **Tavily 官方 API Key**：登录管理面板后添加到 Key 池里，代理会自动挑选并转发到 Tavily 上游。

参考项目：`xuncv/TavilyProxyManager`：<https://github.com/xuncv/TavilyProxyManager>

---

## 🚀 功能特性

- **透明代理**：完整转发至 `https://api.tavily.com`（支持所有路径与方法）。
- **Master Key 鉴权**：客户端通过 `Authorization: Bearer <MasterKey>` 安全访问。
- **智能 Key 池管理**：
  - 优先使用剩余额度最高的 Key。
  - 同额度 Key 随机打散，有效防止请求过于集中触发频率限制。
- **自动故障切换**：遇到 `401` / `429` / `432` / `433` 等错误时，自动尝试 Key 池中的下一个可用 Key。
- **MCP 支持**：内置 HTTP MCP (Model Context Protocol) 端点，可轻松接入 Claude、VS Code 等 AI 工具。
- **可视化管理面板**：
  - **Key 管理**：便捷添加、删除及同步多个 Tavily Key 的额度信息。
  - **用量统计**：通过图表直观展示请求量与额度消耗趋势。
  - **请求日志**：详细记录每次请求，支持过滤筛选与手动清理。
- **自动化任务**：每月 1 号自动重置额度，定期清理历史日志。
- **开箱即用**：Go 二进制单文件部署，内嵌 Web UI（Vite + Vue 3 + Naive UI）。

---

## 🛠️ 环境要求

- **Docker / Docker Compose** (推荐部署方式，无需本地环境)
- **Go**: `1.23+` & **Node.js**: `20+` (仅用于本地手动编译)

---

## 📦 快速部署 (Docker)

直接使用 GHCR 镜像部署，**无需本地编译**。

### 1. 使用 Docker Compose (推荐)

创建 `docker-compose.yml` 文件：

```yaml
version: "3.8"
services:
  tavily-proxy:
    image: ghcr.io/xuncv/tavilyproxymanager:latest
    container_name: tavily-proxy
    ports:
      - "8080:8080"
    environment:
      - LISTEN_ADDR=:8080
      - DATABASE_PATH=/app/data/proxy.db
      - TAVILY_BASE_URL=https://api.tavily.com
      - UPSTREAM_TIMEOUT=30s
    volumes:
      - ./data:/app/data
      - /etc/localtime:/etc/localtime:ro
    restart: unless-stopped
```

执行启动：

```bash
docker-compose up -d
```

### 2. 使用 Docker 原生命令

```bash
docker run -d \
  --name tavily-proxy \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e DATABASE_PATH=/app/data/proxy.db \
  ghcr.io/xuncv/tavilyproxymanager:latest
```

### 3. 使用 Render Blueprint（一键部署，免费实例）

仓库根目录已提供 `render.yaml`，推送到 GitHub 后即可直接在 Render 中按 Blueprint 部署。

1. 将仓库推送到 GitHub（建议使用公开仓库，或给 Render GitHub App 授权私有仓库访问）。
2. 打开 Render 控制台，选择 **New > Blueprint**。
3. 选择你的仓库并批准 `render.yaml` 中的服务配置。
4. 选择 **Free** 实例类型。
5. 首次创建时，Render 会提示填写 `ADMIN_PASSWORD`；`ADMIN_USERNAME` 默认为 `admin`，因此首次登录管理面板通常是 `admin / 你填写的密码`。
6. 部署完成后，访问 Render 分配的 `onrender.com` 地址。

Render 部署说明：

- `render.yaml` 现在默认使用 **Free** 实例类型，适合预览、演示和轻量自用。
- Render Free Web Service **不支持 Persistent Disk**，而当前项目使用 SQLite；因此服务重启、重新部署或 Free 实例被平台重建后，运行期数据可能丢失。
- 可能丢失的数据包括：已添加的 Tavily Key、请求日志、缓存、Master Key、以及保存在数据库中的自定义设置。
- 由于 `ADMIN_USERNAME` / `ADMIN_PASSWORD` 由环境变量初始化，所以只要你在 Render 中保留这些环境变量，管理面板账号密码通常仍然可预测；但其他数据库内容不能保证保留。
- Render Free Web Service 约 15 分钟无流量后会休眠，下一次访问会有冷启动延迟。
- 如果你需要稳定持久化数据，建议后续改用付费实例并挂载 Persistent Disk，或将存储层从 SQLite 升级为外部数据库。
- `render.yaml` 将 `autoDeployTrigger` 设为 `off`，更适合 README 中的“一键部署”场景，避免你后续 push 代码时把所有通过按钮创建的实例都自动重部署。如果你部署的是自己的 fork，可在 Render 面板中改为 `On Commit`。

如果你之后修改了仓库名，请同步替换 Deploy to Render 按钮中的仓库地址：

```md
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/yunshenwuchuxun/TavilyProxy)
```

---

## 🔑 首次运行：登录与鉴权

首次启动后，你通常需要先完成三件事：登录管理面板、添加 Tavily Key、保存 Master Key。

### 1. 登录管理面板

- 对于**全新数据库**，如果你没有在首次启动前设置 `ADMIN_USERNAME` / `ADMIN_PASSWORD`，默认账号密码是 `admin / admin`。
- 如果你在首次启动前设置了 `ADMIN_USERNAME` / `ADMIN_PASSWORD`，则首次登录使用你设置的值。
- 管理员凭据会在初始化后持久化到数据库中；后续再修改环境变量，**不会覆盖**已有数据库中的管理员账号密码。

### 2. 获取 Master Key

服务在**首次启动**时会自动生成一个随机的 **Master Key**。它用于后续代理 REST API 和 MCP 调用，**不用于管理面板登录**。

如果你部署在 Render Free 实例上，请特别注意：由于数据库默认不持久化，`Master Key` 可能会在服务重建或重新部署后变化。

您可以通过以下命令查看控制台日志来获取它：

```bash
docker logs tavily-proxy 2>&1 | grep "master key"
```

**日志示例：**
`time=2026-03-08T15:16:50.725+08:00 level=INFO msg="generated master key" master_key=your_generated_master_key_here`

### 3. 首次使用建议顺序

1. 打开管理面板并使用管理员账号密码登录。
2. 在 Key 管理页面添加一个或多个 Tavily 官方 API Key。
3. 在设置页保存好 `Master Key`。
4. 使用 `Master Key` 调用 `/search`、`/extract`、`/crawl` 或 `/mcp`。

> **提示**：请将 Master Key 用于 API 客户端调用，并将管理员用户名/密码单独保存用于管理面板登录。

---

## 🛠️ 本地开发与手动编译

如果您需要修改源码并自行构建：

1.  **启动后端**:
    ```bash
    go run ./server
    ```
2.  **启动前端**:
    ```bash
    cd web && npm install && npm run dev
    ```

默认情况下，从仓库根目录执行 `go run ./server` 会把数据写入 `server/data/app.db`。
Docker 示例则使用挂载到 `/app/data/proxy.db` 的 `./data`，因此本地手动运行和 Docker 运行默认不会共享同一套凭据，除非您显式指定同一个数据库文件。

Render Free 实例没有 Persistent Disk，因此它更接近“临时环境 / 预览环境”，而不是长期稳定运行环境。

如果你想在本地模拟 Render 的数据目录，可以显式指定：

```bash
DATABASE_PATH=./data/proxy.db go run ./server
```

**手动编译二进制产物**:

- **Windows**: `.\scripts\build_all.ps1`
- **Linux/macOS**: `./scripts/build_all.sh`

**使用 Dockerfile 本地构建镜像**:

```bash
docker build -t my-tavily-proxy .
```

---

## 📖 使用指南

### REST API 代理

客户端调用方式与 Tavily 官方 API 完全一致，只需将 API 地址替换为代理地址，并使用 **Master Key**：

```bash
curl -X POST "http://localhost:8080/search" \
  -H "Authorization: Bearer <MASTER_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"query": "最新 AI 技术趋势", "search_depth": "basic"}'
```

**兼容性说明**:

- 支持 `{"api_key": "<MASTER_KEY>"}` 或 `{"apiKey": "<MASTER_KEY>"}`。
- 支持 GET 参数 `?api_key=<MASTER_KEY>`。

### MCP (Model Context Protocol)

服务在 `http://localhost:8080/mcp` 提供 HTTP MCP 端点。

默认启用无状态模式（`MCP_STATELESS=true`），可避免客户端出现 `session not found`。
如需有状态会话，请将 `MCP_STATELESS=false`，并确保上游反向代理正确透传 `Mcp-Session-Id` 且启用会话粘性（sticky）。

#### VS Code 配置示例 (配合 mcp-remote)

```json
{
  "servers": {
    "tavily-proxy": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote",
        "http://localhost:8080/mcp",
        "--header",
        "Authorization: Bearer YOUR_MASTER_KEY"
      ]
    }
  }
}
```

---

## ⚙️ 配置项 (环境变量)

| 变量名             | 说明                 | 默认值                   |
| :----------------- | :------------------- | :----------------------- |
| `LISTEN_ADDR`      | 服务监听地址         | 空；未设置时回退到 `PORT` |
| `PORT`             | 监听端口（Render 会自动注入） | `8080`          |
| `DATABASE_PATH`    | SQLite 数据库路径    | `./server/data/app.db`（源码运行） |
| `TAVILY_BASE_URL`  | 上游 Tavily API 地址 | `https://api.tavily.com` |
| `UPSTREAM_TIMEOUT` | 上游请求超时时间     | `150s`                   |
| `MCP_STATELESS`    | MCP 是否无状态模式   | `true`                   |
| `MCP_SESSION_TTL`  | MCP 会话空闲超时     | `10m`                    |
| `ADMIN_USERNAME`   | 管理面板用户名       | `admin`                  |
| `ADMIN_PASSWORD`   | 管理面板密码         | `admin`（建议生产环境覆盖） |
| `ADMIN_SESSION_TTL`| 管理员会话有效期     | `24h`                    |
| `LOG_LEVEL`        | 日志级别             | `info`                   |

> Docker 镜像默认会将 `DATABASE_PATH` 设置为 `/app/data/proxy.db`；当前 Render Free Blueprint 不挂载磁盘，因此数据库使用容器内的临时文件系统。

---

## 📄 开源协议

本项目基于 MIT 协议开源，完整文本见 `LICENSE`。
