.PHONY: build run test migrate

build:
	go build -o bin/gophermart cmd/gophermart/main.go

run:
	go run cmd/gophermart/main.go

test:
	go test ./...

migrate:
	psql $$DATABASE_URI -f migrations/001_init.sql

docker-run:
	docker-compose up --build