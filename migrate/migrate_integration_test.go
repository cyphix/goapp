//go:build integration

package migrate_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/migrate"
	"github.com/cyphix/goapp/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }

func TestLifecycle(t *testing.T) {
	cfg := testdb.Config(t)
	db, err := database.Open(&cfg.Database)
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	sqlDB, err := db.DB()
	require.NoError(t, err)
	status := func(want uint) {
		t.Helper()
		version, dirty, err := migrate.New(testfixtures.FS, "").Status(sqlDB)
		require.NoError(t, err)
		require.Equal(t, want, version)
		require.False(t, dirty)
		require.NoError(t, sqlDB.Ping(), "caller pool closed")
	}
	status(0)
	require.NoError(t, migrate.New(testfixtures.FS, "").Up(sqlDB))
	status(1)
	require.NoError(t, migrate.New(testfixtures.FS, "").Down(sqlDB, 1))
	status(0)
	require.NoError(t, migrate.New(testfixtures.FS, "").Up(sqlDB))
	status(1)
	_, err = sqlDB.Exec("UPDATE public.schema_migrations SET dirty = true")
	require.NoError(t, err)
	version, dirty, err := migrate.New(testfixtures.FS, "").Status(sqlDB)
	require.NoError(t, err)
	require.Equal(t, uint(1), version)
	require.True(t, dirty)
	require.Error(t, migrate.New(testfixtures.FS, "").Up(sqlDB), "dirty database accepted")
	require.NoError(t, migrate.New(testfixtures.FS, "").Force(sqlDB, 1))
	status(1)
	require.Error(t, migrate.New(testfixtures.FS, "").Down(sqlDB, 0), "zero steps accepted")
	require.NoError(t, migrate.New(testfixtures.FS, "").Reset(sqlDB))
	status(1)
	require.NoError(t, migrate.New(testfixtures.FS, "public").Reset(sqlDB))
	status(1)
	require.Error(t, migrate.New(testfixtures.FS, "").Force(sqlDB, -2), "invalid version accepted")
	_, err = sqlDB.Exec(`CREATE SCHEMA "quoted""schema"`)
	require.NoError(t, err)
	require.NoError(t, migrate.New(testfixtures.FS, `quoted"schema`).Reset(sqlDB))
	status(1)
	require.NoError(t, sqlDB.Close())
	require.Error(t, migrate.New(testfixtures.FS, "").Up(sqlDB), "closed pool accepted")
}

func TestCustomSchema(t *testing.T) {
	db := testdb.Open(t)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(1)
	for _, name := range []string{"custom_schema", `quoted"schema`} {
		require.NoError(t, migrate.New(testfixtures.FS, name).Up(pool))
		version, dirty, err := migrate.New(testfixtures.FS, name).Status(pool)
		require.NoError(t, err)
		require.False(t, dirty)
		require.NotZero(t, version)
		require.NoError(t, migrate.New(testfixtures.FS, name).Reset(pool))
	}
	var path string
	require.NoError(t, pool.QueryRow("SHOW search_path").Scan(&path))
	require.Equal(t, `"public"`, path)
}

func TestFilesystemFailures(t *testing.T) {
	db := testdb.Open(t)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(1)
	for _, runner := range []*migrate.Runner{
		migrate.New(nil, ""),
		migrate.New(os.DirFS(t.TempDir()+"/missing"), ""),
		migrate.New(fstest.MapFS{"000001_bad.up.sql": &fstest.MapFile{Data: []byte("INVALID SQL;")}}, "broken"),
	} {
		require.Error(t, runner.Up(pool), "invalid filesystem/migration accepted")
		require.NoError(t, pool.Ping(), "caller pool closed")
	}
	empty := migrate.New(fstest.MapFS{}, "empty")
	version, dirty, err := empty.Status(pool)
	require.NoError(t, err)
	require.Zero(t, version)
	require.False(t, dirty)
}
