# ─── 构建阶段 ───
FROM golang:1.25-alpine AS builder

WORKDIR /src

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /fwalizer .

# ─── 运行阶段 ───
FROM alpine:3.20

# 安全：创建非 root 用户
RUN adduser -D -g '' appuser

COPY --from=builder /fwalizer /usr/local/bin/fwalizer

USER appuser

# Alpine 不含 pidof，使用 killall -0 检测进程存活
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD killall -0 fwalizer || exit 1

ENTRYPOINT ["/usr/local/bin/fwalizer"]
