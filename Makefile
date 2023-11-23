
bin/viambuildhat: go.mod *.go cmd/module/*.go
	go build -o bin/viambuildhat cmd/module/cmd.go

lint:
	gofmt -s -w .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test

module: bin/viambuildhat
	tar czf module.tar.gz bin/viambuildhat

all: test module 


