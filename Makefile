.PHONY: build run test clean help

build:
	@go run ./tools/build build

run:
	@go run ./tools/build run

test:
	@go run ./tools/build test

clean:
	@go run ./tools/build clean

help:
	@go run ./tools/build help 

