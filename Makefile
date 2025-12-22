.PHONY: build run clean test docker docker-up docker-down

build:
	go build -o bin/socket-service ./cmd/main.go

run:
	go run ./cmd/main.go

clean:
	rm -rf bin/

test:
	go test -v ./...

docker:
	docker build -t cdex-socket-service .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

tidy:
	go mod tidy

lint:
	golangci-lint run

dev:
	air
