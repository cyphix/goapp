# goapp

Shared PostgreSQL application infrastructure for Cyphix Go services. Go 1.26+.
No app models, domain data, HTTP server, logging middleware, GraphQL, frontend or
compose files live here.

## Consumption

Add this repository as a git submodule at `pkg/goapp`, require
`github.com/cyphix/goapp`, and set:

```go
replace github.com/cyphix/goapp => ./pkg/goapp
replace github.com/cyphix/goauth => ./pkg/goauth
replace github.com/cyphix/gologg => ./pkg/gologg
```

Fresh clones need `git submodule update --init --recursive`. Dependency replaces
do not propagate; consumers must supply all three. For standalone development,
clone goauth and gologg as siblings; this module replaces them with `../goauth`
and `../gologg`.

## Public API

- `config`: Shared, ServerConfig, LogConfig, Database, RedisConfig, AuthConfig;
  `(*Shared).Validate() []string`, `(*Database).Validate() []string`,
  `(*Shared).SharedConfig() *Shared`.
  `Load[T any](cfg *T, shared func(*T) *Shared, validate func(*T) []string) ([]string, error)`
  parses optional dotenv and env/v11, aggregates shared/app validation, and returns
  warnings without logging. No caching: apps embed Shared, set display-name and
  app defaults, and own cached Load/Get and Warnings.
- `database.Open(cfg *config.Database) (*gorm.DB, error)`: PostgreSQL only;
  quoted DSN/search path, pool sizing and ping. Caller closes the returned pool.
- `migrate.New(src fs.FS, schema string) *Runner`: root-level golang-migrate SQL.
  Methods `Up(db *sql.DB) error`, `Down(db *sql.DB, steps int) error`,
  `Status(db *sql.DB) (uint, bool, error)`,
  `Force(db *sql.DB, version int) error`, `Reset(db *sql.DB) error`.
  Empty schema means public. Identifiers are quoted; Reset drops with CASCADE.
  Operations restore search_path, leave caller pools open, do not log, and treat
  no-change as success.
- `migratecmd.Main[T Configuration](load func() (T, error), src fs.FS, seed func(context.Context, T, *gorm.DB) error)`:
  full CLI, config errors/warnings, gologg and exit status. Configuration requires
  `SharedConfig() *config.Shared` (promoted by embedding) and
  `ConfigWarnings() []string` (an app method returning Warnings).
  `Run(cfg *config.Shared, src fs.FS, args []string, seed SeedFunc) error`
  is the non-exiting runner; SeedFunc is `func(context.Context, *gorm.DB) error`.
  Commands: up, down [N] (default 1), status, force <version>, reset.
  Reset requires development, admin credentials and a seed hook.
- `redisconn.OpenRedis(cfg *config.RedisConfig) (*redis.Client, error)`:
  pinged Redis, defaults localhost:6379; caller closes it.
- `authsetup.AuthConfig(cfg *config.Shared, db *gorm.DB, rdb *redis.Client) (goauth.Config, error)`
  and `NewAuth(cfg *config.Shared, db *gorm.DB, rdb *redis.Client) (*goauth.Auth, error)`:
  shared goauth setup, development-only auto-login and trusted proxies.
- `httpx.ClientIPExtractor(trustedProxies string) (echo.IPExtractor, error)`:
  explicit CIDR trust only; empty returns nil (goauth's direct-peer default).
- `devseed.Run(ctx context.Context, cfg *config.Shared, db *gorm.DB, seed SeedFunc, options Options) error`:
  owns Redis, bootstraps goauth/init admin, optionally ensures a regular user,
  then invokes the data hook. SeedFunc is `func(context.Context, *gorm.DB) error`.
  `Options{User: &User{Username: "...", Password: "..."}}` opts into the account;
  nil User skips it, nil seed means auth only. Domain data/options stay in apps.
- `testdb.Main(m *testing.M, src fs.FS) int`,
  `Config(t testing.TB) *config.Shared`, `Open(t testing.TB) *gorm.DB`:
  call Main from TestMain with app migrations. Uses TEST_DATABASE_URL or an
  ephemeral testcontainers PostgreSQL server per binary. Each Config/Open gets
  an isolated database dropped via t.Cleanup; Open applies the supplied FS.
  Docker unavailability skips DB tests; invalid explicit URLs/connections fail.
- `dbquery.QuietNotFound(db *gorm.DB) *gorm.DB`: GORM scope for expected missing
  rows, keeping debug diagnostics and all other parent logger behavior.
- `make/migrate.mk`: include from the app Makefile for migrate-up, migrate-down
  (N optional), migrate-status and migrate-reset. MIGRATE_CMD defaults to
  `go run ./cmd/migrate`; override it if an app entrypoint differs.

An app migration entrypoint can be only:

```go
func main() {
    migratecmd.Main(config.Load, migrations.FS, devseed.Run)
}
```

App devseed.Run closes over its configuration when passing the domain hook.
Healtho, Assethoarder, Mystica and Basilisk skip the regular development user;
Nymph, Entrack, Finacialfalcon and Homelab opt in.

## Verification

```sh
gofmt -l .
go build ./...
go vet ./...
go test ./...
TEST_DATABASE_URL=postgres://... go test -p 1 -tags=integration ./...
```

Integration fixtures contain only a tiny items table, not an application's schema.
PostgreSQL only. Shared defaults/env names have config tests; display name is
caller-owned. Migration/reset guards belong in migratecmd, never app startup.

## Secret scanning

BetterLeaks gates PRs and pushes to main with a digest-pinned workflow.
Local hooks follow goauth's conventions and are opt-in per clone:

```sh
./scripts/install-git-hooks.sh
./scripts/betterleaks-scan.sh
```

Extend .betterleaks.toml only for documented fixture false positives.
Never weaken detection or commit real .env files or credentials.
