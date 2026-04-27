FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gym-log .

FROM scratch
WORKDIR /app
COPY --from=builder /app/gym-log .
EXPOSE 8080
ENV ADDR=:8080
ENV DB_PATH=/data/gym.db
VOLUME ["/data"]
ENTRYPOINT ["./gym-log"]
