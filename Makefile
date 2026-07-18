.PHONY: build run docker-build docker-run clean test

# ─── 本地开发 ───

# 编译
build:
	go build -ldflags="-s -w" -o fwalizer .

# 本地运行（需要先创建 .env 文件）
run: build
	./fwalizer

# 运行测试
test:
	go test -v -race ./...

# 代码检查
lint:
	go vet ./...

# ─── Docker ───

DOCKER_IMAGE ?= fwalizer

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	docker run -d --env-file .env --name fwalizer --restart=always $(DOCKER_IMAGE)

docker-stop:
	docker stop fwalizer || true
	docker rm fwalizer || true

docker-logs:
	docker logs -f fwalizer

# ─── 清理 ───
clean:
	rm -f fwalizer
	docker stop fwalizer 2>/dev/null || true
	docker rm fwalizer 2>/dev/null || true
