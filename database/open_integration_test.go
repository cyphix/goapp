//go:build integration

package database_test

import (
	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestOpenPostgres(t *testing.T) {
	cfg := testdb.Config(t)
	cfg.Database.Schema = `quoted"schema`
	db, err := database.Open(&cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := pool.Ping(); err != nil {
		t.Fatal(err)
	}
	if got := pool.Stats().MaxOpenConnections; got != 100 {
		t.Fatalf("max connections=%d", got)
	}
	var path string
	if err := pool.QueryRow("SHOW search_path").Scan(&path); err != nil {
		t.Fatal(err)
	}
	if path != `"quoted""schema"` {
		t.Fatalf("search_path=%q", path)
	}
}
