APP_NAME = subscriptions
CONFIG = config.yaml

up: docker-up build migrate-up run

docker-up:
	docker compose up -d

run:
	go run ./cmd/main.go

build:
	go build -o $(APP_NAME) ./cmd/main.go

tidy:
	go mod tidy

swag:
	swag init --generalInfo cmd/main.go --output docs

migrate-up:
	goose -dir ./migrations postgres "host=localhost user=postgres password=postgres dbname=subscriptions sslmode=disable" up

migrate-down:
	goose -dir ./migrations postgres "host=localhost user=postgres password=postgres dbname=subscriptions sslmode=disable" down
