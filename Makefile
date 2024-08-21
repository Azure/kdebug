build:
	CGO_ENABLED=0 go build -o bin/kdebug github.com/nevembalint/kdebug/cmd
	CGO_ENABLED=0 go build -o bin/run-as-host github.com/nevembalint/kdebug/cmd/run-as-host

build-win:
	CGO_ENABLED=0 GOOS=windows go build -o bin/kdebug.exe github.com/nevembalint/kdebug/cmd

test:
	CGO_ENABLED=0 go test -v github.com/nevembalint/kdebug/...
