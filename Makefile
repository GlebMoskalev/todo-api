ifneq (,$(wildcard .env))
include .env
export
endif

env_prepare:
	cp -n .env.example .env

migrate_up:
	migrate -path migrations -database "postgres://$(DB_HOST)/$(DB_NAME)?sslmode=disable" up

migrate_down:
	migrate -path migrations -database "postgres://$(DB_HOST)/$(DB_NAME)?sslmode=disable" down