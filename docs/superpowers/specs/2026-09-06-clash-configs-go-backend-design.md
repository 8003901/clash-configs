# clash-configs 后端 Go 重构设计

- 日期：2026-09-06
- 状态：已确认，待实现
- 主题：把后端从 Kotlin/Spring Boot 4 忠实 1:1 移植为 Go（Gin + GORM + 纯 Go SQLite）

## 背景

当前后端在 `backend/`，技术栈为 Kotlin 2.2 + Spring Boot 4 + Hibernate/JPA + H2 文件库 + Spring Security + GraalVM native-image。它实现了一个自托管的 Clash 订阅合并工具，核心能力有三块：

1. **订阅配置管理** `ClashConfig`：URL + 名称 + 更新周期（DAY/WEEK）+ 拉取的 YAML 内容 + 流量信息，带每日/每周定时自动刷新。
2. **合并配置** `ClashConfigsMerge`：名称 + token + 合并结果（JSON 模板），多对多关联订阅配置。
3. **合并引擎** `ConfigController.processConfig`：业务核心，把模板 JSON 与多个订阅的 proxies 合并，做 filter-key 过滤、去空组、rules 校验、流量头汇总，最终输出 YAML 给 Clash 客户端。

外加单用户认证（admin / 默认 password，bcrypt）与静态文件托管（Vue 构建产物打进二进制）。

## 目标

1. 新建目录 `backend-go/`，用 Go 重写后端，旧 `backend/` 目录保留不动。
2. **API 路径、JSON 字段名（camelCase）、请求/响应语义、鉴权行为、合并引擎逻辑完全一致**，现有 Vue 前端零改动即可对接。
3. Go 编译为单一静态二进制，取代 GraalVM native-image 的部署诉求。
4. 数据从 H2 迁移到 SQLite，迁移现有数据，bcrypt 密码哈希原样复用（登录行为不变）。

## 关键决策

| 决策点 | 结论 |
| --- | --- |
| 移植定位 | 忠实 1:1 移植（API/行为一致，前端不改） |
| 目录 | 新目录 `backend-go/`，旧 `backend/` 保留 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | `glebarez/sqlite`（纯 Go、无 CGO），替代 H2 |
| 密码 | `golang.org/x/crypto/bcrypt`（与 Spring BCrypt 兼容） |
| YAML | `gopkg.in/yaml.v3`（`map[string]any` 树，非字节级一致，见「已识别的取舍」） |
| 定时 | `robfig/cron/v3` |
| 会话 | 内存 session + HttpOnly cookie |
| 静态托管 | `embed.FS`（go:embed） |
| 数据迁移 | 借 H2 自带工具 dump 成 SQL/CSV，再由 Go 命令导入 SQLite |

**有意偏离 1:1 的一处修复**：`GET /clash_configs` 原实现因 `Example.of(enabled=true)` 只返回 `enabled=true` 的配置，导致「停用」后配置从列表消失、无法再启用（前端有启用/停用开关，行为矛盾）。本次**顺带修复**为返回全部配置（含停用）。

## 目标目录结构

```
backend-go/
├── go.mod / go.sum
├── Makefile                     # 构建入口：sync 前端产物、build、migrate
├── cmd/
│   ├── server/main.go           # HTTP 服务入口
│   └── migrate/main.go          # H2→SQLite 一次性迁移命令
├── internal/
│   ├── config/config.go         # 配置（环境变量）+ config-template.json 加载
│   ├── model/model.go           # GORM 模型：ClashConfig / ClashConfigsMerge / User
│   ├── store/                   # 数据访问（GORM）
│   ├── service/                 # 业务：clash_config / merge / user
│   ├── merge/merge.go           # 合并引擎（processConfig 核心）
│   ├── web/                     # Gin 路由、handler、鉴权中间件、session
│   └── scheduler/scheduler.go   # 每日/每周定时刷新
├── web/                         # 前端构建产物（go:embed；构建前从 frontend/dist 同步）
├── resources/config-template.json
├── Dockerfile
└── docker-compose.yaml（更新）
```

## 数据模型与迁移

### 表结构（与实体一一对应）

| 表 | 关键字段 |
| --- | --- |
| `clash_configs` | id(UUID string)、url、name、enabled、updateSchedule(DAY/WEEK)、content(长文本)、createdAt、updatedAt、subscriptionUserinfo |
| `clash_configs_merge` | id、name、token、config(长文本 JSON 模板)、createdAt、updatedAt |
| `clash_configs_merge_config` | 多对多中间表：clash_configs_merge_id ↔ config_id |
| `user` | id、username、password(bcrypt) |

- UUID 存 string，时间存 `time.Time`（序列化为 RFC3339 字符串，对齐 Spring 的 `Date` ISO-8601 输出）。
- `updateSchedule` 存字符串常量 `DAY` / `WEEK`。

### 迁移方案（一次性、幂等）

H2 的 `.mv.db` 是私有 MVStore 格式，Go 无成熟读取器。方案：

1. `cmd/migrate` 内部调用 H2 自带工具（`org.h2.tools.Script` 或 `CSVWRITE`，需 JDK 与 h2 jar，可从 Gradle 缓存或下载）把 4 张表 dump 成 SQL/CSV。
2. Go 命令把 dump 结果导入 SQLite（新建 `data/demo.db`）。
3. 幂等：已存在数据则跳过；提供 `--force` 覆盖。
4. bcrypt 哈希原样迁移，`user.password` 直接可用。

## API 与认证契约（逐条对齐）

### 端点

| 方法/路径 | 说明 | 认证 |
| --- | --- | --- |
| `POST /login` | form-encoded username/password，成功建 session + 302 `/` | 公开 |
| `POST /logout` | 销毁 session | 公开 |
| `PUT /users/me/password` | `{oldPassword,newPassword}`，旧密码错误 400 | 需认证 |
| `GET /clash_configs` | 列表（**修复后返回全部**） | 需认证 |
| `POST /clash_configs` | 创建 `{url,name,updateSchedule,enabled}` | 需认证 |
| `GET /clash_configs/{id}?renew=` | 详情，renew=true 时远端刷新 | 需认证 |
| `PUT /clash_configs/{id}` | 更新 | 需认证 |
| `DELETE /clash_configs/{id}` | 删除 | 需认证 |
| `GET /clash_configs_merge` | 列表 | 需认证 |
| `POST /clash_configs_merge` | 创建 `{name,configIds}` | 需认证 |
| `GET /clash_configs_merge/{id}` | 详情 | 需认证 |
| `PUT /clash_configs_merge/{id}` | 更新 `{name,config,configs:[{id}]}` | 需认证 |
| `PUT /clash_configs_merge/{id}/token` | 刷新 token | 需认证 |
| `GET /configs?token=` | 输出合并后的 YAML + `subscription-userinfo` 头 | **公开** |

### JSON 字段名（camelCase，前端契约）

- `ClashConfig`：`id,url,name,enabled,updateSchedule,content,createdAt,updatedAt,subscriptionUserinfo`
- `ClashConfigsMerge`：`id,name,token,config,createdAt,updatedAt,configs`
- 时间字段为 RFC3339 字符串（前端 `new Date(value)` 可解析）。

### 鉴权行为（对齐 Spring Security）

- 公开路径：`/configs/**`、`/login`、`/logout`、`/`、`/index.html`、`/assets/**`、`/favicon.svg`、`/icons.svg`
- 其余需认证。未认证时：
  - AJAX 请求（`X-Requested-With: XMLHttpRequest` 或 `Accept` 含 `application/json`）→ `401` + `{"error":"Unauthorized"}`
  - 否则 → `302 /login`
- 启动时无 admin 则创建（username=admin，password=password，bcrypt）；环境变量 `pwdInit=true` 时重置为默认密码。

## 合并引擎（核心，逐逻辑等价）

`processConfig` 用 `map[string]any` 树重写，步骤等价如下：

1. 读 `resources/config-template.json` 为基础树。
2. 取 `configs` 中「content 非空 且 流量未超 95% 阈值」的订阅，逐个解析其 YAML 的 `proxies`，追加进基础树的 `proxies` 数组。
3. 汇总各订阅 `subscriptionUserinfo`：upload/download/total 求和、expire 取最小；有值则写响应头 `subscription-userinfo: upload=…; download=…; total=…; expire=…`。
4. 收集全部代理名。
5. 遍历 `proxy-groups`：
   - `filter-key == "all"` → 组 proxies 追加全部代理名；
   - `filter-key` 非空 → 按 `|` 拆分关键词，筛选名字 `contains` 任一关键词的代理名追加；
   - 删除 `filter-key` 字段。
6. 移除 proxies 为空的组，并从其它组的 proxies 列表里清掉对它的引用。
7. `checkRules`：每条 rule 按 `,` 拆分，目标索引 = 段数≥3 时取 2、否则取 1；目标代理若非特殊组（DIRECT/REJECT/REJECT-DROP/REJECT-INT/MATCH）且不在组名集合里 → 替换为 `Main Node`。
8. 序列化为 YAML，去掉开头 `---\n`。

**quota 判定**：已用 = upload + download；`used * 100 > total * 95` 时跳过该配置（total ≤ 0 时不判定，保留）。

## 调度 / 静态托管 / 配置

- 调度（robfig/cron，翻译自 Spring cron）：
  - 每日 2:00：`0 0 2 * * ?` → `0 0 2 * * *`（刷新 DAY 配置）
  - 每周一 1:00：`0 0 1 ? * MON` → `0 0 1 * * MON`（刷新 WEEK 配置）
- 静态托管：`frontend/dist` 构建产物同步到 `backend-go/web/`，经 `embed.FS` 打进二进制，路由映射 `/`、`/assets/**`、`/favicon.svg`、`/icons.svg`。
- 配置环境变量：`SERVER_PORT`（默认 8080；Docker 里 80）、SQLite 文件路径（默认 `data/demo.db`）、`pwdInit`。
- 远程订阅拉取：`net/http` + `User-Agent: clash-verge/*`，读取响应头 `subscription-userinfo` 与响应体。

## 部署与测试

- **Dockerfile**：多阶段，`golang:1.23`（`CGO_ENABLED=0`）构建 → `distroless`/`alpine` 运行，单一静态二进制；`VOLUME /data` 挂载 SQLite。
- **docker-compose.yaml**：更新为 `backend-go` 构建上下文，端口映射与 `SERVER_PORT` 对齐原配置。
- **测试**：
  - 合并引擎纯函数，表驱动测试覆盖：filter-key 过滤、去空组、checkRules 替换、quota 跳过、userinfo 汇总、空/非法 YAML 容错。
  - handler 层用 `httptest` + SQLite 临时库做接口级测试（CRUD、鉴权 401/302 分支、密码修改）。
  - 定时调度逻辑用假时钟/直接调用刷新函数验证。

## 已识别的取舍与风险

1. **YAML 键排序**：`yaml.v3` 对 map 输出按字母排序，Jackson 保持插入顺序 → 输出功能等价、Clash 可正常解析，但非字节级一致。已确认接受；若日后需字节一致，改用 `yaml.Node` 保序。
2. **数值类型**：JSON 模板数字经 `encoding/json` 解析为 `float64`，YAML 订阅数字经 `yaml.v3` 解析为 `int`/`float64`；仅影响内部表示，输出数字文本一致（`yaml.v3` 对 `float64(300)` 输出 `300`）。
3. **H2 迁移依赖 JDK**：迁移命令需要一次性 JDK + h2 jar 环境；迁移是一次性操作，不进入运行时依赖。
4. **session 不持久化**：内存 session 与 Spring 默认一致，重启后需重新登录。

## 非目标（YAGNI）

- 不实现多用户/权限体系（保持单 admin）。
- 不改前端，不做前端功能增强。
- 不引入 gRPC/消息队列/Redis 等外部依赖。
- 不追求 YAML 输出字节级一致（见取舍 1）。
- 不把前端构建纳入 Go 模块自动生命周期（由 Makefile 显式同步）。
