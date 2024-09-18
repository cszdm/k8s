# Build the manager binary
FROM golang:1.20.4-alpine3.18 AS builder

WORKDIR /app
# Copy the Go Modules manifests
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o test-pod-maxNum-scheduler main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o test-pod-score-scheduler main.go

FROM alpine:3.12
WORKDIR /app
# COPY --from=builder /app/test-pod-maxNum-scheduler .
COPY --from=builder /app/test-pod-score-scheduler .

# ENTRYPOINT ["./test-pod-maxNum-scheduler"]
ENTRYPOINT ["./test-pod-score-scheduler"]


# 第一阶段，使用 golang 镜像进行应用程序编译
# FROM golang:1.20.4-alpine3.18 AS builder

# 设置工作目录
# WORKDIR /app

# 将项目中的所有文件复制到 /app 目录
# COPY . .

# 编译 Go 应用程序
# RUN go build -o test-pod-maxNum-scheduler .

# 第二阶段，使用更小的基础镜像来运行编译好的程序
# FROM alpine:3.12

# 从 builder 阶段中复制编译好的二进制文件
# COPY --from=builder /app/test-pod-maxNum-scheduler .

# 设置容器启动时要运行的命令
# CMD ["./test-pod-maxNum-scheduler"]