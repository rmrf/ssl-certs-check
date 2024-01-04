default: run_main

run_main:
	go run *.go -config config.toml

build:
	go build ./...

test:
	go test -v ./...

