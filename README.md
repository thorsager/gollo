# gollo
Simple Test web-app in go

[![Go Report Card](https://goreportcard.com/badge/github.com/thorsager/gollo)](https://goreportcard.com/report/github.com/thorsager/gollo)
[![](https://images.microbadger.com/badges/image/thorsager/gollo.svg)](https://microbadger.com/images/thorsager/gollo "Get your own image badge on microbadger.com")
[![](https://images.microbadger.com/badges/version/thorsager/gollo.svg)](https://microbadger.com/images/thorsager/gollo "Get your own version badge on microbadger.com")


## A bit of config
Gollo will bind to the IP expressed in the ENV var `SERVER_IP` on the port expressed in `SERVER_PORT`. if `SERVER_IP` is
not set, Gollo will bin do all available addresses, and if `SERVER_PORT` is not set it will bind to port **8080**.

Also Gollo will print out the message found in the ENV var `GOLLO_MESSAGE` in every response.

If the `DUMP_HEADERS` ENV var is set to a value that evaluates to `true` it will dump request headers
in response as well.

In the same way the `DUMP_ENVIRONMENT` ENV var will dump the current environment in which the "server" is
running.
