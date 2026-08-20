# 构建阶段
FROM golang:1.24-alpine AS builder
WORKDIR /build

# 复制依赖文件（当前目录已是 backend/）
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# 编译（入口在 ./cmd）
RUN go build -ldflags="-w -s" -o app ./cmd

# 运行阶段
FROM alpine:latest
WORKDIR /app

# 复制二进制
COPY --from=builder /build/app .

# 复制配置文件
COPY configs ./configs

EXPOSE 8080
CMD ["./app"]
