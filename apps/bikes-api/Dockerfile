FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY libs/go-auth/go.mod libs/go-auth/go.sum ./libs/go-auth/
COPY libs/go-utils/go.mod libs/go-utils/go.sum ./libs/go-utils/
COPY libs/go-logger/go.mod libs/go-logger/go.sum ./libs/go-logger/
COPY apps/bikes-api/go.mod apps/bikes-api/go.sum ./apps/bikes-api/

WORKDIR /src/apps/bikes-api
RUN go mod download

WORKDIR /src
COPY libs/go-auth ./libs/go-auth
COPY libs/go-utils ./libs/go-utils
COPY libs/go-logger ./libs/go-logger
COPY apps/bikes-api ./apps/bikes-api

WORKDIR /src/apps/bikes-api
RUN go build -ldflags="-s -w -extldflags '-static'" -o /velotrace-bikes ./cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /velotrace-bikes /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY apps/bikes-api/internal/db/sql /migrations
COPY tools/go-prod-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8081

ENTRYPOINT ["/docker-entrypoint.sh", "/velotrace-bikes"]
