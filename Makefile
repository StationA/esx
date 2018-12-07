tools:
	@go install github.com/golang/dep/cmd/dep
	@go get golang.org/x/tools/cmd/stringer

deps: tools
	dep ensure

build: tools
	@go build -o target/esx ./...

install: build
	@go install ./...

target:
	mkdir -p target

release: build target
	@CGO_ENABLED=0 go build -a -o target/esx ./cmd/esx

release-all: build target
	@CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -a -o target/esx.darwin-386 ./cmd/esx
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o target/esx.darwin-amd64 ./cmd/esx
	@CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -o target/esx.linux-386 ./cmd/esx
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o target/esx.linux-amd64 ./cmd/esx
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -a -o target/esx.linux-arm ./cmd/esx
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o target/esx.linux-arm64 ./cmd/esx
	@CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -o target/esx.windows-386.exe ./cmd/esx
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o target/esx.windows-amd64.exe ./cmd/esx

clean:
	@rm -rf target

.PHONY: tools build install release release-all clean
