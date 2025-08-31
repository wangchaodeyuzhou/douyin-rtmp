# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的构建工具和依赖
RUN apk add --no-cache \
    git \
    libpcap-dev \
    gcc \
    musl-dev

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o douyin-rtmp .

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    libpcap \
    ca-certificates \
    tzdata

# 创建非root用户
RUN addgroup -g 1001 app && \
    adduser -D -u 1001 -G app app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/douyin-rtmp .

# 复制配置文件（如果有）
COPY --from=builder /app/configs ./configs

# 更改所有者
RUN chown -R app:app /app

# 切换到非root用户
USER app

# 设置权限
RUN chmod +x douyin-rtmp

# 暴露端口（如果需要）
# EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./douyin-rtmp -version || exit 1

# 启动命令
CMD ["./douyin-rtmp", "-console"]