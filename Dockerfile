# 阶段1：使用 JDK25 构建 Spring Boot JAR
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/graalvm/native-image-community:25 AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.gradle \
    $JAVA_HOME/bin/java -cp gradle/wrapper/gradle-wrapper.jar \
    org.gradle.wrapper.GradleWrapperMain :backend:bootJar --no-daemon

# 阶段2：使用 JDK25 运行 Spring Boot JAR
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/graalvm/native-image-community:25
WORKDIR /
COPY --from=builder /app/backend/build/libs/backend-0.0.1-SNAPSHOT.jar /clash-configs.jar
ENV TZ=Asia/Shanghai
ENV SERVER_PORT=80
VOLUME /tmp
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=70.0", "-Djava.security.egd=file:/dev/./urandom", "-jar", "/clash-configs.jar"]
