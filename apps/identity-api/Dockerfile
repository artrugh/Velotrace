FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src

COPY libs/go-auth/go.mod libs/go-auth/go.sum ./libs/go-auth/

COPY apps/identity-api/go.mod apps/identity-api/go.sum ./apps/identity-api/

WORKDIR /src/apps/identity-api
RUN go mod download

# Copy the rest of the source code
WORKDIR /src
COPY libs/go-auth ./libs/go-auth
COPY apps/identity-api ./apps/identity-api

WORKDIR /src/apps/identity-api
RUN go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /velotrace-identity \
    ./cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /velotrace-identity /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

EXPOSE 8080

ENTRYPOINT ["/velotrace-identity"]
