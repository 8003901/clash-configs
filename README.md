# clash-configs

A self-hosted **Clash subscription merge & configuration center**. It fetches multiple Clash subscription URLs, merges their proxies into a single rule-based config, and exposes one subscription endpoint (`/configs?token=…`) that any Clash client can consume directly.

> 🇨🇳 [中文说明](#中文说明) — a brief Chinese overview is included at the bottom.

## Features

- **Subscription management** (`ClashConfig`)
  - Add / edit / delete / enable-disable subscriptions by URL + name.
  - Update schedule (`DAY` / `WEEK`) with automatic remote refresh (daily at 02:00, weekly on Monday at 01:00).
  - Per-subscription traffic info parsed from the `subscription-userinfo` header.
- **Merge configs** (`ClashConfigsMerge`)
  - Create merges that combine multiple subscriptions (many-to-many), each with a unique token.
  - Editable merge template (JSON) with an in-browser editor.
  - One-click token refresh and copyable subscription link.
- **Merge engine** (the core)
  - Loads the base template (`config-template.json`) and appends the `proxies` of every eligible subscription.
  - `filter-key` group filtering (keyword matching by `|`-separated patterns), empty-group removal, and rule target validation.
  - Aggregates `upload` / `download` / `total` / `expire` into a `subscription-userinfo` response header.
  - Skips subscriptions whose traffic has exceeded 95% of quota.
  - Emits the final merged config as YAML for Clash.
- **Dashboard** — aggregate traffic usage (upload / download / total / earliest expiry).
- **Single-user auth** — `admin` / `password` (bcrypt), with change-password support.

## Architecture

| Directory | Description | Status |
| --- | --- | --- |
| [`frontend/`](frontend/) | Web admin UI — Vue 3 + Vite + TypeScript + shadcn-vue + Tailwind CSS | active |
| [`backend-go/`](backend-go/) | Backend — Go + Gin + GORM + SQLite (pure Go, no CGO), single static binary | **active / canonical** |
| [`backend/`](backend/) | Original backend — Kotlin + Spring Boot 4 + H2 + GraalVM native-image | kept for reference |

The Go backend is a faithful 1:1 rewrite of the Kotlin/Spring Boot backend: the API paths, JSON field names, auth behavior, and merge-engine logic are identical, so the Vue frontend works against either one unchanged. The single static Go binary also embeds the compiled frontend (`go:embed`), replacing the need for a GraalVM native image.

```
clash-configs/
├── frontend/            # Vue 3 admin UI
│   └── src/             # api / components / composables / router / views
├── backend-go/          # Go backend (active)
│   ├── cmd/
│   │   ├── server/      # HTTP server entrypoint
│   │   └── migrate/     # one-time H2 → SQLite migration command
│   └── internal/
│       ├── config/      # env config + embedded config-template.json
│       ├── merge/       # merge engine (processConfig core)
│       ├── model/       # GORM models
│       ├── scheduler/   # daily / weekly auto-refresh
│       ├── service/     # business logic (config / merge / user)
│       ├── store/       # GORM data access
│       └── web/         # Gin routes, handlers, auth middleware, embedded static
├── backend/             # Kotlin + Spring Boot backend (legacy)
└── docs/                # design specs & implementation plans
```

## Tech stack

- **Backend (Go):** Go 1.26, [Gin](https://github.com/gin-gonic/gin), [GORM](https://gorm.io/) + [glebarez/sqlite](https://github.com/glebarez/sqlite) (pure-Go SQLite), [robfig/cron](https://github.com/robfig/cron), `golang.org/x/crypto/bcrypt`, `gopkg.in/yaml.v3`.
- **Backend (Kotlin, legacy):** Kotlin 2.2, Spring Boot 4.0, Hibernate/JPA, H2, Spring Security, GraalVM native-image.
- **Frontend:** Vue 3, Vite 8, TypeScript, shadcn-vue, Tailwind CSS 4, vue-router, axios, CodeMirror.

## Quick start (Docker)

The easiest way to run is with Docker Compose, which builds and runs the Go backend (frontend is already embedded):

```bash
docker compose up -d --build
```

- Admin UI: http://localhost:8780
- Default credentials: `admin` / `password`
- SQLite data persisted to `./data/demo.db` (mounted as `/data`).

## Build & run from source

### Prerequisites

- **Go ≥ 1.26** (backend)
- **Node.js ≥ 20.19** (frontend; Vite 8 requirement) and npm
- **JDK 25** (only for the legacy Kotlin backend)

### Backend (Go)

```bash
cd backend-go

# 1. Sync the built frontend into the embedded static directory
make sync           # copies ../frontend/dist → internal/web/static

# 2. Build (or run directly)
make build          # → bin/clash-configs
make run            # go run ./cmd/server

# 3. Tests
make test
```

The server listens on port `8080` by default (configurable via `SERVER_PORT`).

### Frontend (development)

```bash
cd frontend
npm install
npm run dev         # http://localhost:5173
```

The Vite dev server proxies `/api/*` to the backend (default target `https://config.jijuzhilian.com`, overridable via `PROXY_TARGET`; see `npm run dev:local` / `dev:remote`).

### Legacy backend (Kotlin + Spring Boot)

```bash
./gradlew :backend:bootRun
```

`bootRun` first builds the frontend and syncs it into `backend/src/main/resources/static`, then starts Spring Boot on port `8080`.

## Configuration

The Go backend is configured entirely via environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `SERVER_PORT` | `8080` | HTTP listen port (Compose uses `8780`) |
| `DB_PATH` | `data/demo.db` | Path to the SQLite database file |
| `pwdInit` | `false` | When `true`, resets the admin password to the default on startup |
| `TZ` | (system) | Timezone (Compose sets `Asia/Shanghai`) |

The frontend reads `VITE_BACKEND_URL` to build the public subscription link (`/configs?token=…`); set it to your deployment domain in production.

## API reference

All business APIs require an authenticated session (cookie), except the login/logout and the public subscription endpoint.

| Method & path | Description | Auth |
| --- | --- | --- |
| `POST /login` | Form login (`username` / `password`), sets session cookie | public |
| `POST /logout` | Destroy session | public |
| `PUT /users/me/password` | Change password (`{ oldPassword, newPassword }`) | required |
| `GET /clash_configs` | List subscriptions (all, incl. disabled) | required |
| `POST /clash_configs` | Create `{ url, name, updateSchedule, enabled }` | required |
| `GET /clash_configs/{id}` | Get detail; `?renew=true` forces a remote refresh | required |
| `PUT /clash_configs/{id}` | Update | required |
| `DELETE /clash_configs/{id}` | Delete | required |
| `GET /clash_configs_merge` | List merges | required |
| `POST /clash_configs_merge` | Create `{ name, configIds }` | required |
| `GET /clash_configs_merge/{id}` | Get detail | required |
| `PUT /clash_configs_merge/{id}` | Update `{ name, config, configs: [{ id }] }` | required |
| `PUT /clash_configs_merge/{id}/token` | Refresh merge token | required |
| `GET /configs?token=…` | Public merged config — returns YAML + `subscription-userinfo` header | public |

JSON fields are `camelCase`; timestamps are RFC3339 strings. `updateSchedule` is `"DAY"` or `"WEEK"`.

## Data migration (H2 → SQLite)

The Go backend stores data in SQLite, whereas the legacy Kotlin backend used H2. A one-time migration path is provided:

1. Dump the H2 database to CSV (requires a JDK and the H2 jar):

   ```bash
   backend-go/scripts/h2-dump.sh <h2-jar> backend/data ./dump
   ```

2. Import the CSVs into SQLite:

   ```bash
   cd backend-go
   go run ./cmd/migrate -dump ./dump -out data/demo.db          # fails safely if target exists
   go run ./cmd/migrate -dump ./dump -out data/demo.db -force   # overwrite target
   ```

bcrypt password hashes are migrated as-is, so existing logins keep working.

## Default credentials

- Username: `admin`
- Password: `password`

The admin account is created automatically on first startup if none exists. **Change the password after first login** (Dashboard → 修改密码).

## License

No `LICENSE` file is present in this repository. Assume all rights reserved unless the owner adds one.

---

## 中文说明

`clash-configs` 是一个自托管的 **Clash 订阅合并与配置中心**：拉取多个 Clash 订阅 URL，把它们的代理节点合并进一套基于规则（rule-based）的配置，并对外提供一个订阅接口（`/configs?token=…`），Clash 客户端直接订阅即可。

**核心能力**

- **订阅管理**：URL + 名称 + 更新周期（每天 / 每周，定时自动刷新）+ 启用开关 + 流量信息。
- **合并配置**：多对多关联多个订阅，每个合并配置有独立 token，可编辑模板、刷新 token、复制订阅链接。
- **合并引擎**：基于 `config-template.json` 模板合并各订阅的 `proxies`，做 `filter-key` 关键词过滤、去空组、规则校验、流量头汇总（`subscription-userinfo`），超过 95% 流量的订阅自动跳过，最终输出 Clash 可用的 YAML。
- **仪表盘**：订阅流量用量汇总（上传 / 下载 / 总量 / 最早到期）。
- **单用户认证**：默认账号 `admin` / `password`（bcrypt），支持修改密码。

**目录结构**：`frontend/`（Vue 3 管理界面）、`backend-go/`（Go 后端，当前主力，Gin + GORM + SQLite，单静态二进制并内嵌前端产物）、`backend/`（Kotlin + Spring Boot 4 原版，保留参考）。

**快速开始**：`docker compose up -d --build`，访问 http://localhost:8780（默认账号 `admin` / `password`）。
