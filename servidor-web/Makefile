SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := test

.PHONY: test clean

test:
	@set -e 
	trap 'docker compose down -v' EXIT
	
	docker compose down -v
	docker compose build test
	docker compose up --wait db

	docker compose run --rm test sh -c \
		'sqlc generate && go build ./... && go test ./... -v'
	
clean:
	docker compose down -v