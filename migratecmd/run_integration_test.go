//go:build integration

package migratecmd_test

import (
	"context"
	"errors"
	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/migratecmd"
	"github.com/cyphix/goapp/testdb"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestRunCommands(t *testing.T) {
	cfg := testdb.Config(t)
	cfg.Server.Env = "development"
	cfg.Auth.InitAdminUsername, cfg.Auth.InitAdminPasswd = "admin", "test-password"
	for _, args := range [][]string{{"up"}, {"up"}, {"status"}, {"down"}, {"up"}, {"force", "1"}} {
		if err := migratecmd.Run(cfg, testfixtures.FS, args, nil); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	calls := 0
	seed := func(ctx context.Context, db *gorm.DB) error {
		calls++
		return db.WithContext(ctx).Exec("INSERT INTO items (label) VALUES (?)", "seeded").Error
	}
	if err := migratecmd.Run(cfg, testfixtures.FS, []string{"reset"}, seed); err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(&cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	var count int
	if err := pool.QueryRow("SELECT count(*) FROM items").Scan(&count); err != nil || count != 1 || calls != 1 {
		t.Fatalf("%d %d %v", count, calls, err)
	}
	wantErr := errors.New("seed failed")
	if err := migratecmd.Run(cfg, testfixtures.FS, []string{"reset"}, func(context.Context, *gorm.DB) error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatal(err)
	}
}
