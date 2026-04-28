FROM golang:1.23-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gym-log .

FROM scratch
WORKDIR /app
COPY --from=builder /app/gym-log .
COPY --from=builder /app/static ./static
EXPOSE 1118
ENV ADDR=:1118
ENV DB_PATH=/data/gym.db
VOLUME ["/data"]
ENTRYPOINT ["./gym-log"]
