FROM golang:1.23-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gym-log .

# 原版映像源（海外）：FROM gcr.io/distroless/static-debian12
FROM m.daocloud.io/gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/gym-log .
COPY --from=builder /app/static ./static
EXPOSE 1118
ENV ADDR=:1118
ENV USER_DATA_DIR=/data
VOLUME ["/data"]
ENTRYPOINT ["./gym-log"]
