# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
ARG APP_NAME
RUN go build -o main ./cmd/${APP_NAME}/main.go

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 50051 50052 8080
CMD ["./main"]