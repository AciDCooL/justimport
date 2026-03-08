.PHONY: build clean run test lint

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o justimport ./cmd/justimport

clean:
	rm -f justimport

run:
	go run ./cmd/justimport

test:
	go test -v -race ./...

lint:
	golangci-lint run
