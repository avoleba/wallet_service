FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o wallet-api .

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/wallet-api .

EXPOSE 8080

ENV HTTP_ADDR=:8080
ENV DATABASE_DSN=file:/data/wallet.db?_pragma=foreign_keys(1)

VOLUME /data

CMD ["./wallet-api"]
