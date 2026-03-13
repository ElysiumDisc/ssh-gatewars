.PHONY: build run clean vet dev

build:
	go build -o gatewars ./cmd/server

run: build
	./gatewars

dev: build
	./gatewars --seed 42 --planets 20

clean:
	rm -f gatewars gatewars.db

vet:
	go vet ./...
