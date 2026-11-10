//go:build integration

package database_test

import (
	"os"
	"testing"

	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestOpenPostgres(t *testing.T) {
	cfg := testdb.Config(t)
	cfg.Database.Schema = `quoted"schema`
	db, err := database.Open(&cfg.Database)
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	require.NoError(t, pool.Ping())
	require.Equal(t, 100, pool.Stats().MaxOpenConnections)
	var path string
	require.NoError(t, pool.QueryRow("SHOW search_path").Scan(&path))
	require.Equal(t, `"quoted""schema"`, path)
}
