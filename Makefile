BINARY ?= edge
PREFIX ?= /usr/local

.PHONY: build test tidy fmt install clean

build:
	go build -o $(BINARY) ./cmd/edge

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./pkg

install: build
	install -m 0755 $(BINARY) $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
