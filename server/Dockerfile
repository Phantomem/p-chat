# Build stage
FROM golang:1.23.5-bullseye AS builder

WORKDIR /build

COPY . .

RUN go mod download -x

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags '-linkmode=external' -o /server

CMD ["/server"]

## Final stage
#FROM alpine:3.18
#
#WORKDIR /app
#
#COPY --from=builder /server /server
#COPY --from=builder .env .env
#
#EXPOSE 8080 9090
#CMD ["/app/server"]