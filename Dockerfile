FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o votes .

FROM debian:latest

RUN apt-get update
RUN apt install -y ca-certificates

WORKDIR /app

COPY --from=builder /app/votes .

COPY assets assets

COPY .env .env

COPY data/data.db /app/data/data.db

EXPOSE 8080

CMD ["./votes"]
