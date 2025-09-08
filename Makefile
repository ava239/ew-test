compose-up:
	docker compose up --remove-orphans

build:
	docker compose build

compose-down:
	docker compose down -v --remove-orphans

test:
	go test -v ./pkg/api

install-gen:
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

regen-api:
	go tool oapi-codegen --config=oapi-codegen.yaml openapi.yaml

