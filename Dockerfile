FROM golang:latest as builder
WORKDIR /app
COPY . /app/
RUN CGO_ENABLED=0 GOOS=linux go build -a -o terra-monitors ./cmd/terra-monitors/*.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/terra-monitors /app

EXPOSE 8080
ENTRYPOINT [ "/app/terra-monitors" ]
