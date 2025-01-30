FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go get -u github.com/pressly/goose

COPY . .

RUN go build -o url-shortener ./cmd/url-shortener

CMD ["./url-shortener"]