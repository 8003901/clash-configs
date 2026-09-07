# 阶段1：构建
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.26-alpine AS builder
WORKDIR /app
# 国内 Go 模块代理，加速 go mod download
ENV GOPROXY=https://goproxy.cn,direct
# 先复制依赖清单以利用层缓存
COPY go.mod go.sum ./
# 用 BuildKit 缓存挂载持久化模块缓存，避免每次构建重新下载
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /clash-configs ./cmd/server

# 阶段2：运行（纯静态二进制，distroless 亦可）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/gcr.io/iguazio/alpine:3.20
WORKDIR /
COPY --from=builder /clash-configs /clash-configs
ENV TZ=Asia/Shanghai
ENV SERVER_PORT=80
VOLUME /data
ENTRYPOINT ["/clash-configs"]
