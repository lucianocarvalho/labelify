FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETARCH
ARG TARGETOS
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /app/labelify ./cmd/api

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash labelify

RUN mkdir -p /etc/labelify && chown -R labelify:labelify /etc/labelify

WORKDIR /app

COPY --from=builder /app/labelify /bin/labelify

USER labelify

EXPOSE 8080

ENTRYPOINT [ "/bin/labelify" ]
CMD        [ "--config.file=/etc/labelify/config.yaml" ]