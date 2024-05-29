FROM golang:1.21-alpine as builder

ARG REVISION

RUN mkdir -p /temporal-example/

WORKDIR /temporal-example

COPY . .

RUN go mod download

RUN go build -o temporal-ex .

FROM alpine:3.19

ARG BUILD_DATE
ARG VERSION
ARG REVISION

LABEL maintainer="pinylin"

RUN addgroup -S app \
    && adduser -S -G app app \
    && apk --no-cache add \
    ca-certificates curl netcat-openbsd

WORKDIR /home/app

COPY --from=builder /temporal-example/temporal-ex .
COPY --from=builder /temporal-example/config.yaml .

RUN chown -R app:app ./

USER app

CMD ["./temporal-ex"]
