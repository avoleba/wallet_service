FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o wallet-api .

FROM alpine:3.19

WORKDIR /app

ARG TARGETARCH
RUN apk add --no-cache wget tar \
  && arch="${TARGETARCH:-amd64}" \
  && wget -q "https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-${arch}.tar.gz" -O - | tar xz -C /usr/local/bin \
  && apk del wget tar

COPY --from=builder /app/wallet-api .
COPY migrations /migrations
COPY scripts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENV HTTP_ADDR=:8080
ENV DATABASE_DSN=postgres://localhost:5432/wallet?sslmode=disable

VOLUME /data

ENTRYPOINT ["/app/entrypoint.sh"]
