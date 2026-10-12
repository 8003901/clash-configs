FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:22-alpine3.22 AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.26-alpine AS backend-builder
WORKDIR /app/backend-go
ENV GOPROXY=https://goproxy.cn,direct
COPY backend-go/go.mod backend-go/go.sum ./
RUN go mod download
COPY backend-go/ ./
COPY --from=frontend-builder /app/frontend/dist ./internal/web/static
RUN CGO_ENABLED=0 go build -o /clash-configs ./cmd/server

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/gcr.io/iguazio/alpine:3.20
WORKDIR /
COPY --from=backend-builder /clash-configs /clash-configs
ENV TZ=Asia/Shanghai
ENV SERVER_PORT=8080
VOLUME /data
ENTRYPOINT ["/clash-configs"]
