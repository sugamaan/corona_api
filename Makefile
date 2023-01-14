.PHONY: build

up:
	docker compose -f docker-compose-local.yml up

run:
	GO_ENV=local go run main.go

build:
	sam build

deploy:
	sam deploy

local:
	sam local start-api

lint:
	golangci-lint run

errcheck:
	errcheck ./...