FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/app ./cmd/url-shortener/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bin/app /bin/app

COPY --from=builder /app/configs/local.yaml /app/configs/local.yaml
COPY --from=builder /app/migrations /app/migrations

RUN chmod +x bin/app
CMD ["/bin/app"]