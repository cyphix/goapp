# Versioned schema management; reset bootstraps auth and seeds development data.
MIGRATE_CMD ?= go run ./cmd/migrate

.PHONY: migrate-up migrate-down migrate-status migrate-reset

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down $(N)

migrate-status:
	$(MIGRATE_CMD) status

migrate-reset:
	$(MIGRATE_CMD) reset
