FROM golang:1.26

WORKDIR /app

COPY . .

WORKDIR /app/api

RUN go mod tidy

CMD ["go", "run", "main.go"]