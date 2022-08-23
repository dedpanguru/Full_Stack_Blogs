FROM golang:1.18-alpine

WORKDIR /backend

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o backend main.go

CMD ["./backend"]