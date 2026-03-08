# Tavily Proxy & Management Dashboard

[简体中文](./README.md) | English

> The links below assume the repository path will be `yunshenwuchuxun/TavilyProxy`. If you publish under a different repo name, update the links accordingly.

[![Build](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/build-multi-platform.yml/badge.svg)](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/build-multi-platform.yml)
[![Docker Publish](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/yunshenwuchuxun/TavilyProxy/actions/workflows/docker-publish.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Node.js](https://img.shields.io/badge/Node.js-20%2B-339933?logo=nodedotjs&logoColor=white)](https://nodejs.org/)
[![Render Blueprint](https://img.shields.io/badge/Render-Blueprint-46E3B7?logo=render&logoColor=black)](./render.yaml)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/yunshenwuchuxun/TavilyProxy)

A transparent reverse proxy for the Tavily API that aggregates multiple Tavily API Keys into a single **Master Key**. It features a built-in Web UI for managing keys, monitoring usage, and inspecting request logs.

> **Quick distinction**
> - **Dashboard login**: the first login defaults to `admin / admin` if you did not set `ADMIN_USERNAME` / `ADMIN_PASSWORD` before first startup and the database does not already contain saved admin credentials.
> - **Master Key**: used for `/search`, `/extract`, `/crawl`, `/mcp`, and other proxied calls. It is **not** used for dashboard login.
> - **Official Tavily API Keys**: add them in the dashboard after logging in; the proxy will use them for upstream Tavily requests.

Reference project: `xuncv/TavilyProxyManager`: <https://github.com/xuncv/TavilyProxyManager>

---

## 🚀 Features

- **Transparent Proxy**: Seamlessly forwards requests to `https://api.tavily.com` (supports all endpoints/methods).
- **Master Key Authentication**: Secure access via `Authorization: Bearer <MasterKey>`.
- **Intelligent Key Pooling**:
  - Prioritizes keys with the highest remaining quota.
  - Randomly distributes requests among keys with equal quota to prevent rate limiting.
- **Automatic Failover**: Automatically retries with the next available key upon receiving `401`, `429`, `432`, or `433` errors.
- **MCP Support**: Built-in HTTP MCP (Model Context Protocol) endpoint for easy integration with AI tools (e.g., Claude, VS Code).
- **Comprehensive Dashboard**:
  - **Key Management**: Add, delete, and sync quotas for multiple Tavily keys.
  - **Usage Statistics**: Visualized charts for request volume and quota consumption.
  - **Request Logs**: Detailed logs with filtering and manual cleanup options.
- **Automated Tasks**: Monthly quota resets and periodic log cleaning.
- **Self-Contained**: Single binary deployment with embedded Web UI (Vite + Vue 3 + Naive UI).

---

## 🛠️ Requirements

- **Docker / Docker Compose** (Recommended deployment method, no local environment needed)
- **Go**: `1.23+` & **Node.js**: `20+` (Only for manual builds)

---

## 📦 Quick Deployment (Docker)

Deploy directly using the GHCR image, **no local compilation required**.

### 1. Using Docker Compose (Recommended)

Create a `docker-compose.yml` file:

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

Start the service:

```bash
docker-compose up -d
```

### 2. Using Docker CLI

```bash
docker run -d \
  --name tavily-proxy \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e DATABASE_PATH=/app/data/proxy.db \
  ghcr.io/xuncv/tavilyproxymanager:latest
```

### 3. Deploy with Render Blueprint (one-click, Free instance)

The repository root now includes a `render.yaml`, so after you push the repo to GitHub you can deploy it directly on Render with a Blueprint.

1. Push this repository to GitHub.
2. In the Render Dashboard, choose **New > Blueprint**.
3. Select your repository and approve the service defined in `render.yaml`.
4. Choose the **Free** instance type.
5. During the initial creation flow, Render will prompt you for `ADMIN_PASSWORD`; `ADMIN_USERNAME` defaults to `admin`, so the first dashboard login is usually `admin / <your password>`.
6. Once the deploy is live, open the generated `onrender.com` URL.

Render-specific notes:

- `render.yaml` now defaults to the **Free** instance type, which is suitable for previews, demos, and lightweight personal use.
- Render Free web services **do not support Persistent Disks**, and this project currently uses SQLite. That means runtime data can be lost whenever the service restarts, redeploys, or is rebuilt by the platform.
- Data that can be lost includes added Tavily keys, request logs, cache contents, the generated Master Key, and database-backed settings.
- Because `ADMIN_USERNAME` / `ADMIN_PASSWORD` are initialized from environment variables, the dashboard login usually remains predictable as long as you keep those env vars configured in Render; other database data does not.
- Free Render web services spin down after about 15 minutes of inactivity, so the next request may experience a cold start delay.
- If you need durable state, move to a paid instance with a Persistent Disk later, or upgrade the app to use an external database instead of SQLite.
- `render.yaml` sets `autoDeployTrigger: off`, which is safer for public “Deploy to Render” flows so future pushes to your repo do not redeploy every instance created from the button. If you are deploying your own fork, you can switch Auto-Deploy to `On Commit` in Render after creation.

If you later rename the repository, update the Deploy to Render button URL as well:

```md
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/yunshenwuchuxun/TavilyProxy)
```

---

## 🔑 First Run: Login and Authentication

After the first startup, you typically need to do three things: sign in to the dashboard, add Tavily keys, and save the Master Key.

### 1. Sign in to the dashboard

- For a **fresh database**, if you did not set `ADMIN_USERNAME` / `ADMIN_PASSWORD` before first startup, the default dashboard credentials are `admin / admin`.
- If you did set `ADMIN_USERNAME` / `ADMIN_PASSWORD` before first startup, use those values for the first login.
- Admin credentials are persisted to the database after initialization, so changing those environment variables later will **not override** an existing database.

### 2. Obtain the Master Key

The service automatically generates a random **Master Key** during its **first startup**. Use this key for proxied REST API and MCP calls. It is **not** used for dashboard login.

If you deploy on a Render Free instance, note that the `Master Key` can change after a rebuild or redeploy because the database is not persisted by default.

You can retrieve it by checking the container logs:

```bash
docker logs tavily-proxy 2>&1 | grep "master key"
```

**Log Example:**
`time=2026-03-08T15:16:50.725+08:00 level=INFO msg="generated master key" master_key=your_generated_master_key_here`

### 3. Recommended first-use flow

1. Open the dashboard and sign in with the admin username/password.
2. Add one or more official Tavily API keys in Key Management.
3. Save the `Master Key` from the settings page.
4. Use the `Master Key` for `/search`, `/extract`, `/crawl`, or `/mcp`.

> **Tip**: Save the Master Key for API clients, and save the admin username/password separately for dashboard access.

---

## 🛠️ Local Development & Manual Building

If you need to modify the code and build it yourself:

1.  **Start Backend**:
    ```bash
    go run ./server
    ```
2.  **Start Frontend**:
    ```bash
    cd web && npm install && npm run dev
    ```

By default, running `go run ./server` from the repository root stores data in `server/data/app.db`.
The Docker examples use `./data` mounted to `/app/data/proxy.db`, so local manual runs and Docker runs do not share credentials unless you point them at the same database file.

Render Free instances do not provide a Persistent Disk, so treat that deployment mode as a temporary or preview environment rather than a durable production setup.

If you want local data to match the Render-style path layout, you can run:

```bash
DATABASE_PATH=./data/proxy.db go run ./server
```

**Manual Binary Build**:

- **Windows**: `.\scripts\build_all.ps1`
- **Linux/macOS**: `./scripts/build_all.sh`

**Local Image Build with Dockerfile**:

```bash
docker build -t my-tavily-proxy .
```

---

## 📖 Usage Guide

### REST API Proxy

Call the proxy exactly as you would the official Tavily API, simply replacing the API base URL and using your **Master Key**:

```bash
curl -X POST "http://localhost:8080/search" \
  -H "Authorization: Bearer <MASTER_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"query": "Latest AI trends", "search_depth": "basic"}'
```

**Compatibility Notes**:

- Supports `{"api_key": "<MASTER_KEY>"}` or `{"apiKey": "<MASTER_KEY>"}` in JSON bodies.
- Supports the `api_key=<MASTER_KEY>` GET parameter.

### MCP (Model Context Protocol)

The server provides an HTTP MCP endpoint at `http://localhost:8080/mcp`.

Stateless mode is enabled by default (`MCP_STATELESS=true`) to avoid `session not found` errors.
If you need stateful sessions, set `MCP_STATELESS=false` and ensure your reverse proxy forwards `Mcp-Session-Id` and uses sticky sessions.

#### VS Code Configuration (with mcp-remote)

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

## ⚙️ Configuration (Environment Variables)

| Variable           | Description              | Default                  |
| :----------------- | :----------------------- | :----------------------- |
| `LISTEN_ADDR`      | Server listening address | empty; falls back to `PORT` when unset |
| `PORT`             | Listening port (injected automatically by Render) | `8080` |
| `DATABASE_PATH`    | Path to SQLite database  | `./server/data/app.db` for source runs |
| `TAVILY_BASE_URL`  | Upstream Tavily API URL  | `https://api.tavily.com` |
| `UPSTREAM_TIMEOUT` | Upstream request timeout | `150s`                   |
| `MCP_STATELESS`    | Enable stateless MCP mode | `true`                  |
| `MCP_SESSION_TTL`  | Idle timeout for MCP session | `10m`               |
| `ADMIN_USERNAME`   | Dashboard admin username | `admin`                  |
| `ADMIN_PASSWORD`   | Dashboard admin password | `admin` (override in production) |
| `ADMIN_SESSION_TTL`| Admin session lifetime   | `24h`                    |
| `LOG_LEVEL`        | Log level                | `info`                   |

> The Docker image sets `DATABASE_PATH=/app/data/proxy.db` by default. The current Render Free Blueprint does not mount a disk, so the database lives on the container's ephemeral filesystem.

---

## 📄 License

This project is licensed under the MIT License. See `LICENSE` for the full text.
