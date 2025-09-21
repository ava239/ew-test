.PHONY: build

compose-up:
	docker compose -f=./deployments/docker-compose.yml up --remove-orphans

build:
	docker compose -f=./deployments/docker-compose.yml build

compose-down:
	docker compose -f=./deployments/docker-compose.yml down -v --remove-orphans

test:
	go test -v ./internal/transport

install-gen:
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

regen-api:
	go tool oapi-codegen --config=./api/oapi-codegen.yaml ./api/openapi.yaml

