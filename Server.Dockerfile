FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .

# Generate SQLc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
WORKDIR /app/server
RUN sqlc generate
WORKDIR /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o main ./server/.

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
