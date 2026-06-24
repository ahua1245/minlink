FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apk --no-cache add gcc musl-dev sqlite-dev libc6-compat zlib-dev && CGO_ENABLED=1 GOOS=linux go build -o minlink ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/minlink .
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data

RUN apk --no-cache add tzdata musl sqlite-libs libc6-compat && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

EXPOSE 8080

CMD ["./minlink"]