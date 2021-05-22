# gollo
Simple Test web-app in go

[![GitHub language count](https://img.shields.io/github/languages/count/thorsager/gollo)](https://github.com/thorsager/gollo)
[![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/thorsager/gollo)](https://github.com/thorsager/gollo)
[![Go Report Card](https://goreportcard.com/badge/github.com/thorsager/gollo)](https://goreportcard.com/report/github.com/thorsager/gollo)
[![Build Status](https://travis-ci.com/thorsager/gollo.svg?branch=master)](https://travis-ci.com/thorsager/gollo)
[![Docker Pulls](https://img.shields.io/docker/pulls/thorsager/gollo)](https://hub.docker.com/r/thorsager/gollo)

## A bit of config
Gollo will bind to the IP expressed in the ENV var `SERVER_IP` on the port expressed in `SERVER_PORT`. if `SERVER_IP` is
not set, Gollo will bin do all available addresses, and if `SERVER_PORT` is not set it will bind to port **8080**.

Also Gollo will print out the message found in the ENV var `GOLLO_MESSAGE` in every response.

If the `DUMP_HEADERS` ENV var is set to a value that evaluates to `true` it will dump request headers in response as
well.

In the same way the `DUMP_ENVIRONMENT` ENV var will dump the current environment in which the "server" is running.

**!Please note that dumping the entire server runtime environment to everybody that asks may not be the most secure
thing in the world**

### Paths

It is now possible to change the _path_ of the **prometheus** and the **health** endpoints by setting the two env vars
`PROMETHEUS_PATH` and `HEALTH_PATH`, note: paths should be _absolute_ (starting with `/`)

# Running the thing

```bash
docker run --rm \
  -e DUMP_ENVIRONMENT=true \
  -e DUMP_HEADERS=true \
  -e HEALTH_PATH=/health \
  -e PROMETHEUS_PATH=/prometheus \
  thorsager/gollo
```