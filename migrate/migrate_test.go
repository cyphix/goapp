package migrate

import (
	"database/sql"
	"testing"
	"testing/fstest"
)

func TestMissingDatabase(t *testing.T) {
	runner := New(fstest.MapFS{}, "")
	for _, run := range []func(*sql.DB) error{
		runner.Up, runner.Reset,
		func(db *sql.DB) error { return runner.Down(db, 1) },
		func(db *sql.DB) error { return runner.Force(db, 1) },
		func(db *sql.DB) error { _, _, err := runner.Status(db); return err },
	} {
		if err := run(nil); err == nil {
			t.Fatal("missing database accepted")
		}
	}
	if err := runner.Down(nil, 0); err == nil || err.Error() != "migration steps must be positive" {
		t.Fatal(err)
	}
}
