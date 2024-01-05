default: build

run_main:
	go run *.go -config configurations/config.toml

build:
	go build ./...

test:
	go test -v ./...

