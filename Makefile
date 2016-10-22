.PHONY: test build fmt help

test: -deps fmt
	gb test all

build: test
	gb build all

fmt:
	goimports -w src/

-deps:
	go get github.com/constabulary/gb/...
	go get golang.org/x/tools/cmd/goimports
