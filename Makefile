lint:
	golangci-lint run --issues-exit-code 0

test: 
	go test -v -count 1 -race -mod vendor -parallel 10 ./...

build:
	go build -v -mod vendor -ldflags="-s -w" ./...
