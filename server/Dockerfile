# Build stage
FROM golang:1.21-bullseye AS builder
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \

## Install required build dependencies
#RUN apk get --no-cache \
#    git \
#    make \
#    gcc \
#    ca-certificates

WORKDIR /app

# Copy module files first for better caching
COPY go.mod go.sum ./

# Download dependencies with retries and verbose logging
RUN go mod download -x || \
    (sleep 5 && go mod download -x) || \
    (sleep 10 && go mod download -x)

# Copy source code
COPY . .

# Build with flags for Alpine compatibility
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags '-linkmode=external' -o /server

# Final stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates libc6-compat
COPY --from=builder /server /app/server
COPY --from=builder /app/.env /app/.env

EXPOSE 8080 9090
CMD ["/app/server"]