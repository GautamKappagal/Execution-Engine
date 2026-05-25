FROM golang:1.22-bookworm

WORKDIR /src
COPY . .

RUN apt-get update && apt-get install -y --no-install-recommends docker.io ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /src/worker
RUN go mod download
RUN go build -o /usr/local/bin/worker .

ENTRYPOINT ["/usr/local/bin/worker"]
