SHELL := /bin/bash 
BASEDIR = $(shell pwd)

all:
	go build -o httpctl -v .

gotool:
	@-gofmt -w  .
	@-go tool vet . |& grep -v vendor

clean:
	rm -f httpctl

.PHONY: all gotool clean
