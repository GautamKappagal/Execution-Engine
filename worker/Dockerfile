FROM golang:1.26

WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y docker.io

WORKDIR /app/worker

RUN go mod tidy

CMD ["go", "run", "main.go"]