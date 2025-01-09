FROM golang:1.23.4-alpine AS builder

WORKDIR /app
COPY . .

# Generate SQLc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
WORKDIR /app/server
RUN sqlc generate
WORKDIR /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./server/.

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]