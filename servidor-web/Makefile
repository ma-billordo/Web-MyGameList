SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := test

.PHONY: test run api-test clean

test:
	@set -e 
	trap 'docker compose down -v' EXIT
	
	docker compose down -v
	docker compose build test
	docker compose up --wait db

	docker compose run --rm test sh -c \
		'sqlc generate && go build ./... && go test -v ./db/sqlc'
	
# Levanta la API (y la base de datos) en http://localhost:8080.
run:
	docker compose up --build server

# Igual que 'test', pero en vez de db_test.go ejecuta requests.hurl contra la API.
api-test:
	@set -e
	trap 'docker compose down -v' EXIT

	docker compose down -v
	docker compose build server
	docker compose up --wait db
	docker compose up -d --wait server

	docker compose exec server hurl --test requests.hurl

clean:
	docker compose down -v