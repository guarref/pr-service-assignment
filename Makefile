# переменные
DB_DSN := "postgres://postgres:PostgresPass@db:5432/prservice?sslmode=disable"
MIGRATE := migrate -path ./migrations -database $(DB_DSN)

OPENAPI_FILE=api/openapi.yml
OPENAPI_GEN_OUTPUT=internal/web/omodels/api.gen.go
OPENAPI_GEN_PACKAGE=omodels

.PHONY: help migrate-new migrate migrate-down migrate-force generate-openapi run build docker-up docker-down

generate-openapi:
	oapi-codegen -generate types,server -package $(OPENAPI_GEN_PACKAGE) -o $(OPENAPI_GEN_OUTPUT) $(OPENAPI_FILE)

migrate:
	$(MIGRATE) up

migrate-down-all:
	$(MIGRATE) down

run: 
	go run cmd/app/main.go

docker-up:
	docker-compose up --build -d

docker-down: 
	docker-compose down

load-tests:
	k6 run ./k6.js

lint: 
	golangci-lint run
