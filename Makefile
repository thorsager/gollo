.PHONY: build test vet run clean

build:
	CGO_ENABLED=0 go build -a -tags netgo \
		-ldflags "-X main.version=$(shell git describe --tags --dirty --always) -w -extldflags -static" \
		-o gollo .

test:
	go test -v -race ./...

vet:
	go vet ./...

run:
	go run .

clean:
	rm -f gollo
