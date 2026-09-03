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
