FROM golang:1.22-bookworm AS build

WORKDIR /src
COPY . .

WORKDIR /src/api
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/api .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]
