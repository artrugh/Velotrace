FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /velotrace-identity \
    ./cmd/main.go


FROM alpine:latest

WORKDIR /

COPY --from=builder /velotrace-identity /velotrace-identity
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

EXPOSE 8080

CMD ["/velotrace-identity"]