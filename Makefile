.PHONY: build run clean

build:
	go build -o gatewars ./cmd/server/

run: build
	./gatewars --port 2222

clean:
	rm -f gatewars server
