
bin/viamgpsd: go.mod *.go cmd/module/*.go
	go build -o bin/viamgpsd cmd/module/cmd.go

lint:
	gofmt -s -w .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test

module: bin/viamgpsd
	tar czf module.tar.gz bin/viamgpsd

all: test module 


