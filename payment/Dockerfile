FROM golang:1.16-alpine as builder

ADD go.* /go/src/

WORKDIR /go/src/

RUN go mod download

COPY . /go/src/

RUN go build -o main

FROM alpine:3

RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk* \
    && adduser -u 1000 -D -s /bin/sh app \
    && install -d -m 0750 -o app -g app /app

WORKDIR /app

COPY --from=builder /go/src/main /app

USER app

ENTRYPOINT ["/app/main"]
