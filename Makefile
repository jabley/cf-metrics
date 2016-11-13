.PHONY: test build fmt help

build: test
	gb build all

-darwin: test
	env GOOS=darwin GOARCH=amd64 gb build all

-linux: test
	env GOOS=linux GOARCH=amd64 gb build all

release: -darwin -linux

test: -deps fmt
	gb test all

fmt:
	goimports -w src/

clean:
	rm -rf bin/

-deps:
	go get github.com/constabulary/gb/...
	go get golang.org/x/tools/cmd/goimports
