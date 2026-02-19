build:
	go build -o incipit .

install:
	go install .

release:
	mkdir -p dist
	GOOS=darwin  GOARCH=amd64 go build -o dist/incipit-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -o dist/incipit-darwin-arm64  .
	GOOS=linux   GOARCH=amd64 go build -o dist/incipit-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build -o dist/incipit-linux-arm64   .

clean:
	rm -f incipit
	rm -rf dist/

.PHONY: build install release clean
