FROM golang:1.27.1-alpine3.24 AS base
WORKDIR /app
RUN apk add --no-cache git openssh-client \
    && git config --system --add safe.directory /app
RUN addgroup -g 1000 go && adduser -h /home/go -G go -D -g 1000 go
RUN go install github.com/air-verse/air@v1.67.4 && chown -R go:go /go
