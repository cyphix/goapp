// Package migrate runs caller-provided PostgreSQL schema migrations.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Runner struct {
	src    fs.FS
	schema string
}

// New uses SQL migration files at the root of src. An empty schema means public.
func New(src fs.FS, schema string) *Runner {
	if schema == "" {
		schema = "public"
	}
	return &Runner{src: src, schema: schema}
}

func (r *Runner) run(db *sql.DB, fn func(*gomigrate.Migrate) error) (err error) {
	if db == nil {
		return errors.New("database is required")
	}
	if r.src == nil {
		return errors.New("migrations filesystem is required")
	}
	schema := r.schema
	if schema == "" {
		schema = "public"
	}
	source, err := iofs.New(r.src, ".")
	if err != nil {
		return fmt.Errorf("opening migrations: %w", err)
	}
	defer func() { err = errors.Join(err, source.Close()) }()
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("opening migration connection: %w", err)
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	var searchPath string
	if err := conn.QueryRowContext(context.Background(), "SHOW search_path").Scan(&searchPath); err != nil {
		return err
	}
	if _, err := conn.ExecContext(context.Background(), "SELECT set_config('search_path', $1, false)", quoteIdentifier(schema)); err != nil {
		return err
	}
	defer func() {
		_, restoreErr := conn.ExecContext(context.Background(), "SELECT set_config('search_path', $1, false)", searchPath)
		err = errors.Join(err, restoreErr)
	}()
	if _, err := conn.ExecContext(context.Background(), "CREATE SCHEMA IF NOT EXISTS "+quoteIdentifier(schema)); err != nil {
		return err
	}
	// WithInstance.Close also closes the caller's pool. WithConnection owns
	// only this dedicated connection, leaving db usable after every operation.
	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{SchemaName: schema})
	if err != nil {
		return fmt.Errorf("opening migration driver: %w", err)
	}
	m, err := gomigrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("initializing migrations: %w", err)
	}
	err = fn(m)
	if errors.Is(err, gomigrate.ErrNoChange) {
		err = nil
	}
	return err
}

// Up applies all pending migrations without closing db.
func (r *Runner) Up(db *sql.DB) error {
	return r.run(db, func(m *gomigrate.Migrate) error { return m.Up() })
}

// Down rolls back steps migrations. A nonpositive step count is invalid.
func (r *Runner) Down(db *sql.DB, steps int) error {
	if steps <= 0 {
		return errors.New("migration steps must be positive")
	}
	return r.run(db, func(m *gomigrate.Migrate) error { return m.Steps(-steps) })
}

// Status returns zero and clean when no migration has been applied.
func (r *Runner) Status(db *sql.DB) (version uint, dirty bool, err error) {
	err = r.run(db, func(m *gomigrate.Migrate) error {
		var e error
		version, dirty, e = m.Version()
		if errors.Is(e, gomigrate.ErrNilVersion) {
			return nil
		}
		return e
	})
	return
}

// Force sets the recorded version without running migrations.
func (r *Runner) Force(db *sql.DB, version int) error {
	return r.run(db, func(m *gomigrate.Migrate) error { return m.Force(version) })
}

// Reset drops and recreates schema, then applies all pending migrations.
// An empty schema means public.
func (r *Runner) Reset(db *sql.DB) error {
	if db == nil {
		return errors.New("database is required")
	}
	if r.src == nil {
		return errors.New("migrations filesystem is required")
	}
	schema := r.schema
	if schema == "" {
		schema = "public"
	}
	quoted := quoteIdentifier(schema)
	if _, err := db.Exec("DROP SCHEMA IF EXISTS " + quoted + " CASCADE; CREATE SCHEMA " + quoted); err != nil {
		return fmt.Errorf("resetting schema: %w", err)
	}
	return r.Up(db)
}

func quoteIdentifier(name string) string { return `"` + strings.ReplaceAll(name, `"`, `""`) + `"` }
