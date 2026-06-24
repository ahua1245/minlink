FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o minlink ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/minlink .
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data

RUN apt-get update && apt-get install -y --no-install-recommends tzdata && rm -rf /var/lib/apt/lists/* && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

EXPOSE 8080

CMD ["./minlink"]
