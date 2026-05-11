FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
RUN go build -o /app/server ./cmd/${SERVICE}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY configs/ ./configs/

RUN sed -i \
      -e 's/127\.0\.0\.1:3307/mysql:3306/g' \
      -e 's/127\.0\.0\.1:6379/redis:6379/g' \
      -e 's/127\.0\.0\.1:2379/etcd:2379/g' \
      -e 's/127\.0\.0\.1:5672/rabbitmq:5672/g' \
      configs/*.yaml

EXPOSE 8080

CMD ["./server"]