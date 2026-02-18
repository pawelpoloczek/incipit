build:
	go build -o cli-md .

install:
	go install .

release:
	mkdir -p dist
	GOOS=darwin  GOARCH=amd64 go build -o dist/cli-md-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -o dist/cli-md-darwin-arm64  .
	GOOS=linux   GOARCH=amd64 go build -o dist/cli-md-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build -o dist/cli-md-linux-arm64   .

clean:
	rm -f cli-md
	rm -rf dist/

.PHONY: build install release clean
