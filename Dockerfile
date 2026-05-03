FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o config-rabbit config_rabbit.go

CMD ["./config-rabbit"]