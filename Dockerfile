FROM golang:1.23 AS builder

ARG VERSION=dev

COPY . /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 GOOS=linux go build -o bin -ldflags="-X 'main.version=$VERSION'" main.go

FROM alpine:3.20

RUN mkdir -p /opt/swagger-docs-provider
WORKDIR /opt/swagger-docs-provider
RUN mkdir data
COPY --from=builder /go/src/app/bin bin

HEALTHCHECK --interval=10s --timeout=5s --retries=3 CMD wget -nv -t1 --spider 'http://localhost/health-check' || exit 1

ENTRYPOINT ["./bin"]
