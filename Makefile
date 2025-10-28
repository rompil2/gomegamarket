.PHONY: build run test migrate

build:
	go build -o bin/gophermart cmd/gophermart/main.go

run:
	go run cmd/gophermart/main.go -a ":8090" -d "postgresql://postgres_user:postgres_password@localhost:5432/postgres_db" -r "http://localhost:8080"

test:
	go test ./...

migrate:
	goose up 

docker-run:
	docker-compose up --build