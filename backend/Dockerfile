FROM golang:1.23-alpine AS build
ENV GOTOOLCHAIN=auto
# 降低 go mod 拉取偶发网络错误（可改为你司内网 GOPROXY）
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server -ldflags="-s -w" ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
