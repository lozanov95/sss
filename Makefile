include .env
DSN := "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}/${POSTGRES_DB}?sslmode=disable"

.PHONY: target
migrate:
	@migrate -path ./migrations -database ${DSN} up

.PHONY: target
migration:
	migrate create -seq -dir ./migrations -ext .sql ${name}

.PHONY: target
up:
	docker compose up -d --build

.PHONY: target
down:
	docker compose down