# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum（如果存在）
COPY go.mod go.sum* ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 同步依赖
RUN go mod tidy

# 构建二进制文件
# CGO_ENABLED=0 生成静态链接的二进制文件
# -ldflags="-w -s" 减小二进制文件大小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /build/cryptosentinel ./cmd/bot

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates（用于 HTTPS 请求）和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/cryptosentinel .

# 复制配置文件目录结构
COPY --from=builder /build/configs ./configs

# 设置文件所有权
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 设置时区（可通过环境变量覆盖）
ENV TZ=Asia/Shanghai

# 暴露健康检查端口（如果需要）
# EXPOSE 8080

# 运行程序
CMD ["./cryptosentinel"]
