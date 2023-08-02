FROM golang:1.20 as build
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY *.go ./
COPY config/ ./config
COPY firefly/ ./firefly
COPY classifier/ ./classifier
RUN go mod download
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ffiiitc

FROM alpine:latest as release
WORKDIR /app
RUN mkdir -p /app/data
COPY --from=build /src/ffiiitc /app/ffiiitc
EXPOSE 8082
ENTRYPOINT  ["/app/ffiiitc"]
