# Multiarch Docker cross-compiled build.
# Ref: https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/

FROM --platform=$BUILDPLATFORM golang:1.21-alpine as build
WORKDIR /src
COPY internal/ ./internal
COPY go.mod ./
COPY go.sum ./
COPY *.go ./
RUN go mod download
RUN go test ./...
ARG TARGETPLATFORM TARGETARCH TARGETOS
RUN echo "Building for ${TARGETPLATFORM}..."
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/ffiiitc

FROM alpine:latest as release
WORKDIR /app
RUN mkdir -p /app/data
COPY --from=build /out/ffiiitc /app/ffiiitc
EXPOSE 8082
ENTRYPOINT  ["/app/ffiiitc"]
