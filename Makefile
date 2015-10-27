
build:
	go build eos-server.go

test:
	go test ./...

fmt: 
	go fmt ./...

vet:
	go vet ./...

