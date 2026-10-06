FROM golang:1.27.1-alpine3.24 AS base
WORKDIR /app
RUN addgroup -g 1000 go && adduser -h /home/go -G go -D -g 1000 go
RUN go install github.com/air-verse/air@latest && chown -R go:go /go