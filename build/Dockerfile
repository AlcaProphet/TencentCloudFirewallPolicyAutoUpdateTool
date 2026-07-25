# 编译阶段
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -tags docker \
    -ldflags="-s -w -X github.com/alcaprophet/fwalizer/version.Version=${VERSION}" \
    -o /fwalizer .

# 运行阶段
FROM alpine:3.20
RUN adduser -D appuser
COPY --from=builder /fwalizer /usr/local/bin/fwalizer
USER appuser
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:9090/api/health 2>/dev/null || pgrep fwalizer || exit 1
ENTRYPOINT ["fwalizer"]
