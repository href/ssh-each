.PHONY: build lint test

build:
	go build -trimpath -o bin/ssh-each main.go

lint:
	golangci-lint run

test:
	go test -race -coverpkg=./... ./...
