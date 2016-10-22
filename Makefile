.PHONY: test build fmt help

build: test
		gb build all

test: -deps fmt
	gb test all

fmt:
	goimports -w src/

clean:
	rm -rf bin/

-deps:
	go get github.com/constabulary/gb/...
	go get golang.org/x/tools/cmd/goimports
