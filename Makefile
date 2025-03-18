include .env
DSN := "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}/${POSTGRES_DB}?sslmode=disable"

.SILENT:
migrate:
	migrate -path ./migrations -database ${DSN} up
up:
	docker compose up -d --build
down:
	docker compose down