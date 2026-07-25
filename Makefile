.PHONY: build run docker-build docker-run clean test lint sdk-update

# ─── 版本信息 ───
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/alcaprophet/fwalizer/version.Version=$(VERSION)

# ─── 本地开发镜像（可选，取消注释启用） ───
# export GOPROXY := https://mirrors.tencent.com/go/,direct
# export GONOSUMCHECK := *
# export GOFLAGS := -mod=mod

# ─── 本地开发 ───

# 编译
build:
	go build -ldflags="$(LDFLAGS)" -o fwalizer .

# 本地运行（需要先创建 .env 文件）
run: build
	./fwalizer

# 运行测试
test:
	go test -v -race ./...

# 代码检查
lint:
	go vet ./...

# 更新所有云厂商 SDK 到最新版
sdk-update:
	go get -u github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common
	go get -u github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse
	go get -u github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc
	go get -u github.com/alibabacloud-go/swas-open-20200601/v3
	go get -u github.com/alibabacloud-go/ecs-20140526/v7
	go get -u github.com/alibabacloud-go/darabonba-openapi/v2
	go get -u github.com/aliyun/credentials-go
	go mod tidy

# ─── Docker ───

DOCKER_IMAGE ?= fwalizer

docker-build:
	docker build -f build/Dockerfile --build-arg VERSION=$(VERSION) -t $(DOCKER_IMAGE) .

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
