# 阶段1：构建阶段，用 GraalVM 把后端编译成本地可执行文件（native-image）
# 采用 glibc（Oracle Linux 基础镜像），与阶段2的 glibc 运行时保持一致
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/graalvm/native-image-community:17 AS builder
WORKDIR /app
# 复制整个 monorepo（含根 settings.gradle、gradlew、backend/，以及已提交的前端静态资源）
COPY . .
# Gradle 依赖走 aliyun 镜像（已配置在 settings.gradle / build.gradle 的 repositories）。
# 该镜像自带 native-image + gcc；显式指向 GRAALVM_HOME，绕过 toolchain 检测的不确定性。
# 只构建后端 native 镜像，避免触发前端 Node 下载（前端产物已提交在 backend/src/main/resources/static）。
# 直接调用 gradle-wrapper.jar 而非 gradlew 脚本：该 OL 镜像未预装 xargs，而 gradlew 脚本会强制校验 xargs。
# -XX:-UseJVMCICompiler：在 arm64 Mac 上以 Rosetta 模拟构建 amd64 时，GraalVM 的 JIT(JVMCI) 会 SIGSEGV，
# 改用标准 C2 JIT 构建；native-image 的 AOT 编译走独立进程，产物仍是标准 x86_64，不受此开关影响。
ENV GRAALVM_HOME=$JAVA_HOME
RUN --mount=type=cache,target=/root/.gradle $JAVA_HOME/bin/java -XX:-UseJVMCICompiler -cp gradle/wrapper/gradle-wrapper.jar org.gradle.wrapper.GradleWrapperMain :backend:nativeCompile --no-daemon

# 阶段2：运行阶段，轻量级 glibc 运行时镜像（native 镜像需与构建端同为 glibc）。
# 走华为云 SWR 镜像，与仓库里既有 openjdk 镜像同一命名空间（Docker Hub 直连在此网络环境不可用）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/debian:bookworm-slim
WORKDIR /
# 可执行文件名与模块名相同（backend）
COPY --from=builder /app/backend/build/native/nativeCompile/backend /clash-configs
ENV TZ=Asia/Shanghai
RUN echo "Asia/Shanghai" > /etc/timezone
VOLUME /tmp
# 端口默认 80，可通过环境变量 SERVER_PORT 覆盖（Spring Boot 宽松绑定读取）
ENV SERVER_PORT=80
ENTRYPOINT ["/clash-configs"]
