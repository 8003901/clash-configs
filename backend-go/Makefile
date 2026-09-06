.PHONY: sync build test run migrate

# 同步前端构建产物到静态目录（供 go:embed）
sync:
	mkdir -p internal/web/static
	cp -r ../frontend/dist/* internal/web/static/

build: sync
	CGO_ENABLED=0 go build -o bin/clash-configs ./cmd/server

test:
	go test ./...

run: sync
	go run ./cmd/server

migrate: sync
	go run ./cmd/migrate
