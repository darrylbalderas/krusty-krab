# Start from a base image with Go installed
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN go build -o app main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
CMD ["./app"]