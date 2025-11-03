.PHONY: help
help: ## Show help
	@echo 'Usage: make [target]'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
.PHONY: run
run: ## Run the application
	go run cmd/factorit/main.go

.PHONY: build
build: ## Build Factorit binary
	go build -o bin/factorit cmd/factorit/main.go

# Testing
.PHONY: test
test: ## Run tests
	go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Database
.PHONY: infra-up
infra-up: ## Start PostgreSQL
	docker-compose up -d 
	@echo "Waiting for PostgreSQL, Prometheus, Grafana and Postgres Exporter to be ready..."
	@sleep 3

.PHONY: infra-down
infra-down: ## Stop PostgreSQL
	docker-compose down

.PHONY: db-logs
db-logs: ## Show PostgreSQL logs
	docker-compose logs -f postgres

.PHONY: db-shell
db-shell: ## Open psql shell
	docker exec -it observ-postgres psql -U postgres -d observ-db

# Migrations
.PHONY: migrate-up
migrate-up: ## Apply all migrations
	@./scripts/migrate.sh up

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@./scripts/migrate.sh down 1

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create name=add_roles)
	@./scripts/migrate.sh create $(name)

.PHONY: migrate-version
migrate-version: ## Show migration version
	@./scripts/migrate.sh version

# Database seeding
.PHONY: db-seed
db-seed: ## Seed database with 50 test users
	@echo "🌱 Seeding database..."
	@docker exec -i observ-postgres psql -U postgres -d observ-db < scripts/seed_users.sql
	@echo "✓ Database seeded with 50 users"

.PHONY: db-clean
db-clean: ## Clean test data from database
	@echo "🧹 Cleaning test data..."
	@docker exec -i observ-postgres psql -U postgres -d observ-db -c "DELETE FROM users WHERE email LIKE '%@example.com';"
	@echo "✓ Test data cleaned"

.PHONY: db-reset
db-reset: migrate-down migrate-up db-seed ## Reset database (down, up, seed)
	@echo "✓ Database reset complete"

# Setup
.PHONY: setup
setup: db-up migrate-up ## Complete setup
	@echo "✓ Setup complete!"

# Clean
.PHONY: clean
clean: ## Clean artifacts
	rm -rf bin/
	rm -f coverage.out