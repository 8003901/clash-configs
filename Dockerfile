# 阶段1：构建阶段，使用带JDK17的Gradle镜像
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/openjdk:17-jdk-alpine AS builder
WORKDIR /app
# 复制项目源码到工作目录
COPY . .
RUN mkdir  /root/.gradle
# 安装 findutils 以获得 xargs
RUN apk update && apk add --no-cache findutils

# 使用 BuildKit 挂载缓存，挂载 Gradle 缓存目录，避免重复下载依赖
RUN --mount=type=cache,target=/root/.gradle sh /app/gradlew build --no-daemon

# 阶段2：运行阶段，基于轻量级 OpenJDK17 运行时镜像
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/openjdk:17-slim
VOLUME /tmp
WORKDIR /
# 从构建阶段复制生成的jar包，调整路径以匹配你的项目结构
COPY --from=builder /app/build/libs/clash-configs-0.0.1-SNAPSHOT.jar /clash-configs.jar
RUN echo "Asia/Shanghai" > /etc/timezone
ENV JAVA_OPTS="-Dserver.port=80"
ENTRYPOINT [ "sh", "-c", "java -XX:MaxRAMPercentage=70.0 $JAVA_OPTS -Djava.security.egd=file:/dev/./urandom -jar /clash-configs.jar" ]