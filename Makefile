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

# Infrastructure & Docker
.PHONY: docker-up
docker-up: ## Start all infrastructure + factorit app in Docker
	@echo "🚀 Starting complete infrastructure..."
	docker-compose up -d --build
	@echo "⏳ Waiting for services to be ready..."
	@sleep 5
	@echo "✅ Infrastructure ready!"
	@echo ""
	@echo "📊 Services available:"
	@echo "  - Factorit API:        http://localhost:8080"
	@echo "  - Factorit Health:     http://localhost:8080/health"
	@echo "  - Grafana:             http://localhost:3000 (admin/admin)"
	@echo "  - Prometheus:          http://localhost:9090"
	@echo "  - Jaeger UI:           http://localhost:16686"
	@echo "  - Loki API:            http://localhost:3100"
	@echo "  - PostgreSQL:          localhost:5432"
	@echo ""
	@echo "💡 View logs: make logs"

.PHONY: docker-down
docker-down: ## Stop all Docker containers
	@echo "🛑 Stopping all containers..."
	docker-compose down
	@echo "✅ All containers stopped"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart all Docker containers

.PHONY: docker-clean
docker-clean: ## Stop containers and remove volumes (⚠️  deletes data)
	@echo "⚠️  This will delete all data (postgres, loki logs, etc)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "✅ Containers and volumes removed"; \
	else \
		echo "❌ Cancelled"; \
	fi

.PHONY: logs
logs: ## Show logs from all containers
	docker-compose logs -f

.PHONY: logs-app
logs-app: ## Show logs from factorit app only
	docker-compose logs -f factorit

.PHONY: logs-loki
logs-loki: ## Show Loki logs
	docker-compose logs -f loki

.PHONY: logs-promtail
logs-promtail: ## Show Promtail logs
	docker-compose logs -f promtail

.PHONY: infra-up
infra-up: ## Start only infrastructure (without factorit app)
	@echo "🚀 Starting infrastructure services..."
	docker-compose up -d postgres prometheus grafana postgres-exporter jaeger loki promtail
	@echo "⏳ Waiting for services to be ready..."
	@sleep 3
	@echo "✅ Infrastructure ready!"

.PHONY: infra-down
infra-down: ## Stop only infrastructure
	docker-compose stop postgres prometheus grafana postgres-exporter jaeger loki promtail

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
setup: infra-up migrate-up ## Complete setup (infra + migrations)
	@echo "✓ Setup complete!"

.PHONY: setup-full
setup-full: docker-up ## Complete setup with factorit in Docker
	@echo "⏳ Waiting for factorit to be ready..."
	@sleep 3
	@echo "✓ Full setup complete!"

# Clean
.PHONY: clean
clean: ## Clean artifacts
	rm -rf bin/
	rm -f coverage.out