# goapp — Agent Rules

Go 1.26+ shared infrastructure. Apps consume pkg/goapp git submodules with local
replacements. README.md documents the public API.

- Check branch/status, fetch origin, verify dependency siblings before work.
- No app imports, models, domain seeds, server wiring, logging middleware,
  GraphQL, frontend or compose. App options/hooks belong in apps.
- PostgreSQL only, no alternative drivers/tests. Migration inputs are fs.FS.
- config alone loads dotenv/env; return warnings, never log. Apps own caching.
- Use gologg for migratecmd/devseed logs. Migration/database helpers return errors.
- Preserve env names/defaults, reset guards, safe identifiers, connection ownership
  and search_path restoration. migratecmd permits reset only in development.
- Active development: no deprecated aliases, compatibility shims or dual APIs.
- go.mod uses ../goauth and ../gologg sibling replaces, matching goauth.
- Check gofmt -l ., go build ./..., go vet ./..., go test ./..., plus
  TEST_DATABASE_URL=... go test -p 1 -tags=integration ./... for DB changes.
- Tests use tiny module-owned migration fixtures, never app migrations/imports.
- Tear down only processes/containers you start. Helpers clean up owned DBs
  and testcontainers instances.
- BetterLeaks follows goauth. Hook installation is opt-in per clone.
- Never commit .env, credentials, generated reports or dependency caches.
- Commit the module first, then bump app submodule pointers.
