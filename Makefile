MIGRATIONS_DIR := migrations

.PHONY: dev down migrate-up migrate-down migrate-create migrate-force

# Boot Postgres and apply all migrations — one command for a ready local DB.
dev:
	docker compose up -d
	$(MAKE) migrate-up

# Stop the stack (keeps the volume; use `docker compose down -v` to reset data).
down:
	docker compose down

# Apply all pending migrations.
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Roll back the most recent migration.
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

# Scaffold a new migration pair: make migrate-create name=<desc>
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Reset a dirty state to a known-good version: make migrate-force version=<n>
migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(version)
