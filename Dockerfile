FROM golang:1.24-alpine

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o config_rabbit config_rabbit.go

CMD ["./config_rabbit"]