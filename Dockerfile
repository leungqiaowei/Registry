# 阶段 1：构建阶段（使用多阶段构建以减少最终镜像大小）
FROM golang:1.22-alpine AS builder
# 安装构建所需工具
RUN #apk add --no-cache git
# 设置工作目录
WORKDIR /app
# 将 go.mod 和 go.sum 复制到容器中
COPY go.mod ./
# 下载依赖
RUN go mod download
# 将项目代码复制到工作目录中
COPY . .
# 编译 Go 程序（静态编译）
RUN CGO_ENABLED=0 GOOS=linux go  build -o registry ./cmd/registry

# 阶段 2：运行阶段
FROM alpine
# 安装运行时工具（用于调试或检查）
RUN apk add --no-cache bash ca-certificates
# 设置工作目录
WORKDIR /app
# 从构建阶段复制编译好的二进制文件和必要资源
COPY --from=builder /app/registry /app/registry
COPY --from=builder /app/tmp/data /app/tmp/data
# 暴露服务端口
EXPOSE 8081
# 启动应用程序
CMD ["/app/registry"]
