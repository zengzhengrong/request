FROM golang:1.19-alpine as build

# 容器环境变量添加，会覆盖默认的变量值
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV TZ="Asia/Shanghai"
# 设置工作区
WORKDIR /go/release

# 把全部文件添加到/go/release目录
ADD . .

# 编译：把main.go编译成可执行的二进制文件，命名为app
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -tags netgo -ldflags="-s -w" -installsuffix cgo -o app main.go

# FROM scratch as prod
FROM alpine as prod
LABEL maintainer="zengzhengrong"
ENV TZ="Asia/Shanghai"

# 时区纠正
RUN rm -f /etc/localtime \
    && ln -sv /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
# 在build阶段复制可执行的go二进制文件app
COPY --from=build /go/release/app /zurl

# 启动服务
ENTRYPOINT ["/zurl"]