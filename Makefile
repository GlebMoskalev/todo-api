ifneq (,$(wildcard .env))
include .env
export
endif

help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9_-]+:.*#' Makefile | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

env_prepare: # Copy .env.example to .env if .env does not exist.
	cp -n .env.example .env

migrate_up: # Run database migrations to upgrade schema.
	migrate -path migrations -database "postgres://$(DB_HOST)/$(DB_NAME)?sslmode=disable" up

migrate_down: # Run database migrations to downgrade schema.
	migrate -path migrations -database "postgres://$(DB_HOST)/$(DB_NAME)?sslmode=disable" down

test: # Run all tests in the project.
	go test -v ./...

default: help