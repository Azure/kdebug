build:
	CGO_ENABLED=0 go build -o bin/kdebug github.com/Azure/kdebug/cmd

test:
	CGO_ENABLED=0 go test -v github.com/Azure/kdebug/...
