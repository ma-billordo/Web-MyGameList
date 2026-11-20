SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := test

.PHONY: test run clean

# Tests de la capa de datos (TP2)
test:
	@set -e 
	trap 'docker compose down -v' EXIT
	
	docker compose down -v
	docker compose build test
	docker compose up --wait db

	docker compose run --rm test sh -c \
		'sqlc generate && templ generate  && go build ./... && go test -v ./db/sqlc'
	
# Levanta la base y el servidor (http://localhost:8080)
run:
	@set -e
	trap 'docker compose down -v' EXIT
	docker compose down -v
	docker compose up --build server

clean:
	docker compose down -v