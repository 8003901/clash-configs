# Gradle 多模块 monorepo 化 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把后端（Kotlin/Spring Boot）与前端（Vue 3/Vite）重组成根目录下的 Gradle 多模块 monorepo，目录重命名为 `backend/`、`frontend/`，两个 git 仓库合并并保留历史。

**Architecture:** 根目录建 `settings.gradle`（`include 'backend', 'frontend'`）与最小 `build.gradle`；wrapper 上移到根；后端构建脚本不动；前端用 `com.github.node-gradle.node` 插件把 npm 构建包进 Gradle。

**Tech Stack:** Gradle 8.12.1、Kotlin 1.9.25、Spring Boot 3.4.2、Java 17、Node 22.12.0、`com.github.node-gradle.node` 7.0.1、git 2.50（仅用 core git，不引入 `git filter-repo`）。

**Spec:** `docs/superpowers/specs/2026-09-03-gradle-multimodule-monorepo-design.md`

## Global Constraints

- Gradle wrapper 版本 **8.12.1**，`distributionUrl` 保持华为云镜像（`repo.huaweicloud.com/gradle/gradle-8.12.1-bin.zip`）。
- 后端 Java toolchain **17**；Kotlin/Spring Boot/依赖管理插件版本**原样保留**在 `backend/build.gradle`（不改）。
- 前端 Node 版本 **22.12.0**（Vite 8 要求 Node ≥20.19），`download = true`，`distBaseUrl` 走 npmmirror。
- `rootProject.name = 'clash-configs'`；模块目录名 `backend`、`frontend`。
- `pluginManagement`（阿里云 Gradle 插件镜像 + `gradlePluginPortal()`）搬到根 `settings.gradle`。
- 历史保留用 **core git 的 `git mv` + `git merge --allow-unrelated-histories`**；不安装 `git filter-repo`。
- 前端构建**默认不挂到根 `build` 生命周期**（避免后端/Docker 构建意外下载 Node）。
- 两个仓库目前都有**未提交改动**，必须先提交再动 git（Task 1）。

---

### Task 1: 提交两个仓库未提交的改动

**Files:**
- Modify: `clash-configs/`（提交 `SecurityConfig.kt`、`IndexController.kt` 的修改，以及未跟踪的 `src/main/resources/static/`）
- Modify: `clash-configs-web/`（提交 `src/api/auth.ts`、`src/api/http.ts`、`src/lib/config.ts`、`src/router/index.ts` 的修改）

**Interfaces:**
- Consumes: 无。
- Produces: 两个仓库 working tree **clean**，供 Task 2 安全 `git clone`。

- [ ] **Step 1: 提交后端未提交改动**

```bash
cd /Users/keketata/idea-projects/github/clash/clash-configs
git add -A
git status --short
git commit -m "chore: 提交未提交的后端改动"
```

- [ ] **Step 2: 提交前端未提交改动**

```bash
cd /Users/keketata/idea-projects/github/clash/clash-configs-web
git add -A
git status --short
git commit -m "chore: 提交未提交的前端改动"
```

- [ ] **Step 3: 验证两个仓库 clean**

```bash
cd /Users/keketata/idea-projects/github/clash/clash-configs && git status --short --branch
cd /Users/keketata/idea-projects/github/clash/clash-configs-web && git status --short --branch
```

Expected: 两个仓库都显示 `## main`（后端可能带 `[ahead N]`），且 **无** `M`/`??` 行。

---

### Task 2: 合并两个仓库为根 monorepo（保留历史），删除 site/

**Files:**
- Create: 根 git 仓库（`git init -b main`）
- Move: 后端历史整体进 `backend/`；前端历史整体进 `frontend/`
- Delete: `backend/site/`（旧前端）

**Interfaces:**
- Consumes: Task 1 产出两个 clean 仓库。
- Produces: 根 git 仓库，含 `backend/`（无 `site/`）与 `frontend/` 两个目录及完整历史。

- [ ] **Step 1: 准备后端子目录历史（临时 clone）**

```bash
cd /Users/keketata/idea-projects/github/clash
git clone ./clash-configs /tmp/clash-mono-backend
cd /tmp/clash-mono-backend
mkdir backend
find . -mindepth 1 -maxdepth 1 ! -name '.git' ! -name 'backend' -exec mv {} backend/ \;
git add -A
git commit -m "chore: 将后端源码移入 backend/ 子目录"
```

- [ ] **Step 2: 准备前端子目录历史（临时 clone）**

```bash
cd /Users/keketata/idea-projects/github/clash
git clone ./clash-configs-web /tmp/clash-mono-frontend
cd /tmp/clash-mono-frontend
mkdir frontend
find . -mindepth 1 -maxdepth 1 ! -name '.git' ! -name 'frontend' -exec mv {} frontend/ \;
git add -A
git commit -m "chore: 将前端源码移入 frontend/ 子目录"
```

- [ ] **Step 3: 备份原仓库并清空根目录**

```bash
cd /Users/keketata/idea-projects/github/clash
mkdir -p /tmp/clash-original-backup
mv clash-configs /tmp/clash-original-backup/
mv clash-configs-web /tmp/clash-original-backup/
ls -la   # 期望只剩 .DS_Store、.claude/、.vscode/、docs/ 等未跟踪文件
```

- [ ] **Step 4: 初始化根仓库并合并两段历史**

> 先确认 git 身份已配置（后端仓库能提交说明通常已全局配置），否则 `git commit` 会失败：

```bash
cd /Users/keketata/idea-projects/github/clash
git config user.email >/dev/null 2>&1 || echo "WARN: git user.email 未配置，先运行 git config --global user.email/name"
git init -b main
git commit --allow-empty -m "chore: 初始化 monorepo 根仓库"
git remote add backend /tmp/clash-mono-backend
git fetch backend
git merge --allow-unrelated-histories backend/main -m "chore: 合并后端历史"
git remote add frontend /tmp/clash-mono-frontend
git fetch frontend
git merge --allow-unrelated-histories frontend/main -m "chore: 合并前端历史"
```

- [ ] **Step 5: 删除旧前端 site/**

```bash
cd /Users/keketata/idea-projects/github/clash
rm -rf backend/site
git add -A
git commit -m "chore: 删除旧前端 site/"
```

- [ ] **Step 6: 验证**

```bash
cd /Users/keketata/idea-projects/github/clash
ls -la                       # 期望看到 backend/ 和 frontend/
test ! -e backend/site && echo "site removed OK"
git log --oneline --graph -15   # 期望看到 backend 与 frontend 两条历史都合入
git status --short --branch  # 期望 clean（除 docs/ 等未跟踪）
```

Expected: `backend/`、`frontend/` 都在；`backend/site` 不存在；`git log --graph` 能看到两段历史交汇；历史保留（后端原来的 `feat:`/`fix:` 提交、前端的 `feat: initialize...` 提交都可达）。

---

### Task 3: 根 Gradle 文件 + wrapper 上移 + 删 backend settings.gradle

**Files:**
- Create: `settings.gradle`、`build.gradle`、`.gitignore`
- Move: `backend/gradlew` → `gradlew`；`backend/gradlew.bat` → `gradlew.bat`；`backend/gradle/` → `gradle/`
- Delete: `backend/settings.gradle`

**Interfaces:**
- Consumes: Task 2 的根仓库。
- Produces: 根 Gradle 工程，`./gradlew projects` 能列出 `:backend` 与 `:frontend`。

- [ ] **Step 1: 移动 wrapper 到根、删除后端 settings.gradle**

```bash
cd /Users/keketata/idea-projects/github/clash
git mv backend/gradlew gradlew
git mv backend/gradlew.bat gradlew.bat
git mv backend/gradle gradle
git rm backend/settings.gradle
```

- [ ] **Step 2: 创建根 `settings.gradle`**

写入 `settings.gradle`（用 Write 工具，内容如下）：

```groovy
pluginManagement {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/gradle-plugin' }
        gradlePluginPortal()
    }
}

rootProject.name = 'clash-configs'

include 'backend', 'frontend'
```

- [ ] **Step 3: 创建根 `build.gradle`**

写入 `build.gradle`（最小，内容如下）：

```groovy
// 根工程：多模块下几乎为空，公共约定留待需要时再补充
```

- [ ] **Step 4: 创建根 `.gitignore`**

写入 `.gitignore`（仅根级，内容如下）：

```gitignore
# OS
.DS_Store

# IntelliJ IDEA
.idea/
*.iml
*.ipr
*.iws

# VS Code
.vscode/

# Gradle（根）
.gradle/
build/
```

- [ ] **Step 5: 验证多模块被识别**

```bash
cd /Users/keketata/idea-projects/github/clash
./gradlew projects
```

Expected: 输出包含 `Root project 'clash-configs'` 以及 `Project ':backend'`、`Project ':frontend'`。（首次运行会从华为云镜像下载 Gradle 8.12.1，耗时较长属正常。）

- [ ] **Step 6: 提交**

```bash
cd /Users/keketata/idea-projects/github/clash
git add -A
git commit -m "chore: 建立根 Gradle 多模块（settings/build/gitignore + wrapper 上移）"
```

---

### Task 4: 前端接入 node-gradle 插件

**Files:**
- Create: `frontend/build.gradle`

**Interfaces:**
- Consumes: Task 3 的根工程已 `include 'frontend'`。
- Produces: `./gradlew :frontend:npm_run_build` 产出 `frontend/dist`。

- [ ] **Step 1: 创建 `frontend/build.gradle`**

写入 `frontend/build.gradle`（内容如下）：

```groovy
plugins {
    id 'com.github.node-gradle.node' version '7.0.1'
}

node {
    // Vite 8 需要 Node >= 20.19
    version = '22.12.0'
    download = true
    // 国内网络下载 Node 走 npmmirror 镜像；非国内环境可删除此行
    distBaseUrl = 'https://npmmirror.com/mirrors/node'
    // 可选：npm 安装走 npmmirror registry
    // npmInstallCommandArgs = ['--registry=https://registry.npmmirror.com']
}

// 插件不保证 npm_run_build 自动依赖 npmInstall，这里显式挂上
tasks.named('npm_run_build') {
    dependsOn 'npmInstall'
}
```

- [ ] **Step 2: 验证插件任务已注册**

```bash
cd /Users/keketata/idea-projects/github/clash
./gradlew :frontend:tasks --all | grep -iE 'npm|nodeSetup'
```

Expected: 能看到 `nodeSetup`、`npmSetup`、`npmInstall`、`npm_run_build`、`npm_run_dev` 等任务。

- [ ] **Step 3: 执行前端构建**

```bash
cd /Users/keketata/idea-projects/github/clash
./gradlew :frontend:npm_run_build
```

Expected: `BUILD SUCCESSFUL`，且 `frontend/dist/` 生成（含 `index.html`、`assets/`）。首次会从 npmmirror 下载 Node 22.12.0 并 `npm install`，耗时较长属正常。

- [ ] **Step 4: 提交**

```bash
cd /Users/keketata/idea-projects/github/clash
git add frontend/build.gradle
git commit -m "build: 前端接入 node-gradle 插件"
```

---

### Task 5: 调整 Docker 构建上下文与产物路径

**Files:**
- Modify: `backend/Dockerfile`
- Modify: `backend/docker-compose.yaml`

**Interfaces:**
- Consumes: Task 3（wrapper 已在根）、Task 4（前端模块存在，但 Docker 只构建后端）。
- Produces: Docker 能从根上下文构建出 `backend-0.0.1-SNAPSHOT.jar`。

- [ ] **Step 1: 修改 `backend/Dockerfile`**

把构建阶段改为根上下文、精确构建后端、修正产物路径：

```dockerfile
# 阶段1：构建阶段，使用带JDK17的Gradle镜像
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/openjdk:17-jdk-alpine AS builder
WORKDIR /app
# 复制整个 monorepo（含根 settings.gradle、gradlew、backend/）
COPY . .
RUN mkdir /root/.gradle
# change_mirror.sh 现在位于 backend/ 下
RUN sh backend/change_mirror.sh
# 安装 findutils 以获得 xargs
RUN apk update && apk add --no-cache findutils
# 只构建后端，避免触发前端 Node 下载
RUN --mount=type=cache,target=/root/.gradle sh /app/gradlew :backend:build --no-daemon
# 阶段2：运行阶段，基于轻量级 OpenJDK17 运行时镜像
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/openjdk:17-slim
VOLUME /tmp
WORKDIR /
# jar 名随模块名变为 backend-0.0.1-SNAPSHOT.jar
COPY --from=builder /app/backend/build/libs/backend-0.0.1-SNAPSHOT.jar /clash-configs.jar
RUN echo "Asia/Shanghai" > /etc/timezone
ENV JAVA_OPTS="-Dserver.port=80"
ENTRYPOINT [ "sh", "-c", "java -XX:MaxRAMPercentage=70.0 $JAVA_OPTS -Djava.security.egd=file:/dev/./urandom -jar /clash-configs.jar" ]
```

- [ ] **Step 2: 修改 `backend/docker-compose.yaml`**

`build` 段改为根上下文 + 子目录 Dockerfile：

```yaml
services:
  clash-configs:
    image: clash-configs:0.0.1-SNAPSHOT
    build:
      context: ../
      dockerfile: backend/Dockerfile
    ports:
      - "8780:8780"
    volumes:
      - ./data:/data
    container_name: clash-configs
    restart: always
    environment:
      JAVA_OPTS: '-Dfile.encoding=UTF-8 -Dserver.port=8780'
    cpus: "1"
    mem_limit: 128M
```

- [ ] **Step 3: 验证（逻辑检查 + 可选实测）**

```bash
cd /Users/keketata/idea-projects/github/clash
# 先确认后端 jar 名与路径对得上
ls backend/build/libs/   # 期望存在 backend-0.0.1-SNAPSHOT.jar（Task 6 构建后）
# 可选实测（耗时，镜像下载较大，可跳过）：
# docker build -f backend/Dockerfile .
```

Expected: `Dockerfile` 中 `COPY . .` 的上下文为根（含 `settings.gradle`/`gradlew`/`backend/`），`gradlew :backend:build` 目标正确，jar 路径 `backend/build/libs/backend-0.0.1-SNAPSHOT.jar` 与模块名一致。

- [ ] **Step 4: 提交**

```bash
cd /Users/keketata/idea-projects/github/clash
git add backend/Dockerfile backend/docker-compose.yaml
git commit -m "build: 调整 Docker 构建上下文与产物路径"
```

---

### Task 6: 全量验证 + 提交文档

**Files:**
- Commit: `docs/`（设计文档 + 本计划）

**Interfaces:**
- Consumes: Task 3/4/5。
- Produces: 两端构建均通过，仓库 clean。

- [ ] **Step 1: 构建后端**

```bash
cd /Users/keketata/idea-projects/github/clash
./gradlew :backend:build
```

Expected: `BUILD SUCCESSFUL`，`backend/build/libs/backend-0.0.1-SNAPSHOT.jar` 生成（`test.enabled=false`，不跑测试）。

- [ ] **Step 2: 构建前端**

```bash
cd /Users/keketata/idea-projects/github/clash
./gradlew :frontend:npm_run_build
```

Expected: `BUILD SUCCESSFUL`，`frontend/dist/` 生成。

- [ ] **Step 3: 提交文档与收尾**

```bash
cd /Users/keketata/idea-projects/github/clash
git add -A
git status --short
git commit -m "docs: 记录 Gradle 多模块化设计与实现计划"
git log --oneline -10
```

Expected: `git status` clean；`git log` 里既有两段历史，也有本次 monorepo 化的提交。

---

## 回滚预案（如中途失败）

- 原始两个仓库完整保留在 `/tmp/clash-original-backup/`（`clash-configs`、`clash-configs-web`，含各自 `.git`）。
- 若根仓库合并出错，删除根目录新生成的内容，把 `/tmp/clash-original-backup/` 里的两个目录 `mv` 回来即可恢复原状。
- 临时 clone `/tmp/clash-mono-backend`、`/tmp/clash-mono-frontend` 可随时删除。
