SHELL := /bin/bash 
BASEDIR = $(shell pwd)

all: api
api:
	go build -o httpctl -v .

gotool:
	@-gofmt -w  .
	@-go tool vet . |& grep -v vendor

clean:
	rm -f httpctl
	find . -name "[._]*.s[a-w][a-z]" | xargs -i rm -f {}

.PHONY: all gotool clean api
