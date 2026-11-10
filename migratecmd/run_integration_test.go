//go:build integration

package migratecmd_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/migratecmd"
	"github.com/cyphix/goapp/testdb"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestRunCommands(t *testing.T) {
	cfg := testdb.Config(t)
	cfg.Server.Env = "development"
	cfg.Auth.InitAdminUsername, cfg.Auth.InitAdminPasswd = "admin", "test-password"
	for _, args := range [][]string{{"up"}, {"up"}, {"status"}, {"down"}, {"up"}, {"force", "1"}} {
		require.NoError(t, migratecmd.Run(cfg, testfixtures.FS, args, nil), "%v", args)
	}
	calls := 0
	seed := func(ctx context.Context, db *gorm.DB) error {
		calls++
		return db.WithContext(ctx).Exec("INSERT INTO items (label) VALUES (?)", "seeded").Error
	}
	require.NoError(t, migratecmd.Run(cfg, testfixtures.FS, []string{"reset"}, seed))
	db, err := database.Open(&cfg.Database)
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	var count int
	require.NoError(t, pool.QueryRow("SELECT count(*) FROM items").Scan(&count))
	require.Equal(t, 1, count)
	require.Equal(t, 1, calls)
	wantErr := errors.New("seed failed")
	require.ErrorIs(t, migratecmd.Run(cfg, testfixtures.FS, []string{"reset"}, func(context.Context, *gorm.DB) error { return wantErr }), wantErr)
}
