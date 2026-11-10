package migrate

import (
	"database/sql"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestMissingDatabase(t *testing.T) {
	runner := New(fstest.MapFS{}, "")
	for _, run := range []func(*sql.DB) error{
		runner.Up, runner.Reset,
		func(db *sql.DB) error { return runner.Down(db, 1) },
		func(db *sql.DB) error { return runner.Force(db, 1) },
		func(db *sql.DB) error { _, _, err := runner.Status(db); return err },
	} {
		require.Error(t, run(nil), "missing database accepted")
	}
	require.EqualError(t, runner.Down(nil, 0), "migration steps must be positive")
}
