.PHONY: test build fmt help

build: test
	gb build all

-darwin: test
	env GOOS=darwin GOARCH=amd64 gb build all

linux: test
	env GOOS=linux GOARCH=amd64 gb build all

release: -release-dir -darwin.zip -linux.zip

-linux.zip: linux
	cd bin && zip cf-metrics-linux-amd64.zip cf-metrics-linux-amd64 && cd -
	mv bin/cf-metrics-linux-amd64.zip release/

-darwin.zip: -darwin
	cd bin && zip cf-metrics-darwin-amd64.zip cf-metrics-darwin-amd64 && cd -
	mv bin/cf-metrics-darwin-amd64.zip release/

test: -deps fmt
	gb test all

fmt:
	goimports -w src/

clean:
	rm -rf bin/
	rm -rf release/

-deps:
	go get github.com/constabulary/gb/...
	go get golang.org/x/tools/cmd/goimports

-release-dir:
	mkdir release
