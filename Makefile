# Makefile

# переменные
DB_DSN := "postgres://postgres:PostgresPass@localhost:5432/prservice?sslmode=disable"
MIGRATE := migrate -path ./migrations -database $(DB_DSN)

.PHONY: help migrate-new migrate migrate-down migrate-force run build docker-up docker-down

# таргет для создания новой миграции
migrate-new: 
	migrate create -ext sql -dir ./migrations -seq $(NAME)

# применить все миграции
migrate:
	$(MIGRATE) up

# откатить последнюю миграцию
migrate-down: ## Откатить последнюю миграцию
	$(MIGRATE) down 1

# откатить все миграции
migrate-down-all:
	$(MIGRATE) down

# локальный запуск приложения
run: 
	go run cmd/app/main.go

# сборка бинарника
build: ## Собрать бинарник
	go build -o bin/prservice cmd/app/main.go

# поднять docker-compose
docker-up:
	docker-compose up --build -d

## остановить docker-compose
docker-down: 
	docker-compose down

# запуск линтера
lint: 
	golangci-lint run
