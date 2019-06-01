FROM golang:1-alpine3.8 as build-env

RUN apk add --no-cache git
RUN adduser -D -u 10000 build-user
RUN mkdir /build && chown build-user /build
USER build-user

WORKDIR /build
ADD . /build

RUN CGO_ENABLED=0 go build -ldflags "-X main.version=$(git describe --tags --dirty --always)" -o /build/gollo .

FROM alpine:3.8
RUN adduser -D -u 10000 app-user
USER app-user
WORKDIR /

COPY --from=build-env /build/gollo /

EXPOSE 8080

ENTRYPOINT ["/gollo"]
