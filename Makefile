
bin/viamlegobuildhat: go.mod *.go cmd/module/*.go
	go build -o bin/viamlegobuildhat cmd/module/cmd.go

lint:
	gofmt -s -w .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test

module: bin/viamlegobuildhat
	tar czf module.tar.gz bin/viamlegobuildhat

all: test module 


