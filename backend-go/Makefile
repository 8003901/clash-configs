.PHONY: frontend-build sync build test run migrate

# 先构建前端，再同步静态资源到 go:embed 目录
frontend-build:
	cd ../frontend && npm run build

sync: frontend-build
	rm -rf internal/web/static
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
