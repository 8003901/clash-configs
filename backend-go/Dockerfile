# 阶段1：构建
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.26-alpine AS builder
WORKDIR /app
# 先复制依赖清单以利用层缓存
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /clash-configs ./cmd/server

# 阶段2：运行（纯静态二进制，distroless 亦可）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/gcr.io/iguazio/alpine:3.20
WORKDIR /
COPY --from=builder /clash-configs /clash-configs
ENV TZ=Asia/Shanghai
ENV SERVER_PORT=80
VOLUME /data
ENTRYPOINT ["/clash-configs"]
