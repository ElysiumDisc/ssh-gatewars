.PHONY: build run clean vet

build:
	go build -o gatewars ./cmd/server

run: build
	./gatewars

clean:
	rm -f gatewars

vet:
	go vet ./...
