.PHONY: build run test migrate

build:
	go build -o bin/gophermart cmd/gophermart/main.go

run:
	go run cmd/gophermart/main.go -a ":8090" -d "postgresql://postgres_user:postgres_password@localhost:5432/postgres_db" -r "localhost:8080"

test:
	go test ./...

migrate:
	psql $$DATABASE_URI -f migrations/001_init.sql

docker-run:
	docker-compose up --build