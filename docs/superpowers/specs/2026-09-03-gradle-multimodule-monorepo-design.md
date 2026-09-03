# Gradle 多模块 monorepo 化设计

- 日期：2026-09-03
- 状态：已确认，待实现
- 主题：把后端（Kotlin/Spring Boot）与前端（Vue 3/Vite）重组成根目录下的 Gradle 多模块 monorepo

## 背景

当前仓库根目录 `/Users/keketata/idea-projects/github/clash` 下有两个相互独立的工程：

| 目录 | 类型 | 说明 |
| --- | --- | --- |
| `clash-configs/` | Gradle（Kotlin + Spring Boot 3.4.2 + Java 17，Gradle 8.12.1） | 后端，自带 `settings.gradle`、`gradlew`，内部还嵌了一个旧前端 `site/`（Vue + Vuetify，已停更） |
| `clash-configs-web/` | npm（Vue 3.5 + Vite 8 + TS） | 新版前端，独立仓库 |

两者各自是独立 git 仓库；根目录本身不是 git 仓库。

## 目标

1. 在根目录创建 Gradle 初始化文件（`settings.gradle` 等），把前后端组织成两个子模块。
2. 目录重命名为 `backend/`（后端）与 `frontend/`（前端）。
3. 前端（Node.js 工程）用 `com.github.node-gradle.node` 插件包裹，纳入 Gradle 构建。
4. 两个独立 git 仓库合并成一个根 monorepo，并保留历史。

## 关键决策

| 决策点 | 结论 |
| --- | --- |
| 前端模块 | 新版 `clash-configs-web`（不是旧的 `site/`） |
| 前端集成方式 | `com.github.node-gradle.node` 插件包裹 npm 构建 |
| 目录命名 | 重命名为 `backend/`、`frontend/` |
| Git 结构 | 合并成根 monorepo，保留历史（`git filter-repo` + `git merge --allow-unrelated-histories`） |
| 旧前端 `site/` | 删除 |
| `rootProject.name` | `clash-configs`（沿用应用名） |
| Node 版本 | `22.12.0`（Vite 8 要求 Node ≥20.19） |

## 目标目录结构

```
clash/                         # 根 monorepo（rootProject.name = 'clash-configs'）
├── settings.gradle            # pluginManagement + rootProject.name + include
├── build.gradle               # 最小（几乎为空）
├── gradlew / gradlew.bat / gradle/wrapper/   # 从 backend 移上来的 wrapper（Gradle 8.12.1，华为云镜像）
├── .gitignore                 # 合并两份 gitignore
├── docs/                      # 本设计文档所在
├── backend/                   # 原 clash-configs，删 settings.gradle、删 site/
│   ├── build.gradle           # 基本不动
│   ├── Dockerfile             # jar 名改 backend-*.jar
│   ├── docker-compose.yaml    # build context 改根
│   └── src/…
└── frontend/                  # 原 clash-configs-web
    ├── build.gradle           # 新增 node-gradle 插件
    ├── package.json / src/…
```

## 详细改动

### 1. 根 Gradle 文件
- **`settings.gradle`**（新建）：把原 backend `settings.gradle` 的 `pluginManagement`（阿里云 Gradle 插件镜像）搬上来，加：
  ```groovy
  rootProject.name = 'clash-configs'
  include 'backend', 'frontend'
  ```
- **`build.gradle`**（新建）：最小化，几乎为空。
- **wrapper**：`gradlew`、`gradlew.bat`、`gradle/wrapper/*` 从 `backend/` 移到根；`gradle-wrapper.properties` 保持 Gradle 8.12.1 + 华为云 `distributionUrl`。
- **`.gitignore`**（新建，仅根级）：`.DS_Store`、`.idea/`、`*.iml`、`.vscode/`、`.gradle/`、`build/` 等；`backend/.gitignore` 与 `frontend/.gitignore` 保持原样、各自子目录内继续生效（避免锚定路径如 `/data` 语义错乱）。

### 2. backend 模块
- 删除自己的 `settings.gradle`（多模块里只有根有 settings）。
- `build.gradle` 原样保留（Kotlin/Spring Boot/依赖管理插件版本留在子模块）。
- 删除旧前端 `site/`。
- 保留 `src/main/resources/static/`（旧构建产物，本次不动）。

### 3. frontend 模块
- 新建 `build.gradle`：
  ```groovy
  plugins { id 'com.github.node-gradle.node' version '7.1.0' }
  node {
      version = '22.12.0'
      download = true
      // 国内镜像（视网络情况）
      // distBaseUrl = 'https://npmmirror.com/mirrors/node'
  }
  ```
- 插件自动生成 `npm_install`、`npm_run_build`、`npm_run_dev` 等 task；`./gradlew :frontend:npm_run_build` 即执行 `npm run build`，产出 `frontend/dist`。
- **前端构建默认不挂到根 `build` 生命周期**（不 `assemble.dependsOn npm_run_build`），避免后端/Docker 构建时意外下载 Node；需要一条命令全量构建时再显式加。
- 原有纯 npm 开发流程（`cd frontend && npm run dev`）不受影响，Gradle 模块是增量叠加。

### 4. Docker / 部署
- **构建上下文改为根目录**：因为 `settings.gradle` 和 `gradlew` 现在在根，Dockerfile 需要 `COPY . .` 拿到根（含 `backend/`、`gradlew`、`settings.gradle`）。
- `backend/Dockerfile`：
  - 构建命令改为 `sh /app/gradlew :backend:build`（只构建后端，避免触发前端 Node 下载）。
  - jar 拷贝路径改为 `/app/backend/build/libs/backend-0.0.1-SNAPSHOT.jar`（jar 名随模块名变成 `backend`）。
- `backend/docker-compose.yaml`：build `context` 改为根目录 `..`，`dockerfile` 改为 `backend/Dockerfile`。
- `change_mirror.sh` 若引用相对路径需随移动调整（或移到根目录）。

### 5. Git monorepo 合并（保留历史）
1. 对 `clash-configs` 用 `git filter-repo --to-subdirectory-filter backend`，把全部历史路径重写进 `backend/`。
2. 对 `clash-configs-web` 用 `git filter-repo --to-subdirectory-filter frontend`。
3. 在根新建仓库，把两者作为 remote fetch，`git merge --allow-unrelated-histories` 合并两段历史。
4. 删除 `site/` 并提交。
- 备选（丢历史）：根仓库一次性提交全部文件。

## 验证方式

- `./gradlew :backend:build` 构建后端成功。
- `./gradlew :frontend:npm_run_build` 构建前端，产出 `frontend/dist`。
- 后端 `bootRun` 启动正常；前端 `npm run dev` 启动且 `/api` 代理到 `localhost:8080` 正常。

## 非目标（YAGNI）

- 不自动把 `frontend/dist` 拷进后端 `static/`（后续需要再单独做）。
- 不动后端 `src/main/resources/static/` 里那份旧构建产物。
- 不重命名 `frontend/package.json` 的 `name` 字段（避免无谓 churn）。
