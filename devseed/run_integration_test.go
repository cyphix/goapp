//go:build integration

package devseed_test

import (
	"context"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
	"github.com/cyphix/goapp/devseed"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
	"github.com/cyphix/goauth"
	"gorm.io/gorm"
	"net"
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestRunLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name string
		user *devseed.User
		seed bool
		fail bool
	}{
		{name: "auth only"},
		{name: "data hook", seed: true},
		{name: "optional user", user: &devseed.User{Username: "user", Password: "test-password"}, seed: true},
		{name: "hook error", seed: true, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := miniredis.RunT(t)
			host, port, err := net.SplitHostPort(server.Addr())
			if err != nil {
				t.Fatal(err)
			}
			cfg := &config.Shared{}
			cfg.Server.Env = "development"
			cfg.Redis = config.RedisConfig{Host: host, Port: port}
			cfg.Auth = config.AuthConfig{InitAdminUsername: "admin", InitAdminPasswd: "test-password", RPID: "localhost", RPDisplayName: "Example", Origins: []string{"http://localhost:5173"}, PublicBaseURL: "http://localhost:5173"}
			db := testdb.Open(t)
			wantErr := errors.New("seed failed")
			var hook devseed.SeedFunc
			called := false
			ctx := context.WithValue(context.Background(), struct{}{}, "value")
			if tc.seed {
				hook = func(got context.Context, gotDB *gorm.DB) error {
					called = true
					if got != ctx || gotDB != db {
						t.Fatal("hook arguments changed")
					}
					var count int64
					if err := db.Model(&goauth.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil || count != 1 {
						t.Fatalf("admin not bootstrapped before hook: %d %v", count, err)
					}
					if tc.fail {
						return wantErr
					}
					return db.Exec("INSERT INTO items (label) VALUES (?)", "seeded").Error
				}
			}
			err = devseed.Run(ctx, cfg, db, hook, devseed.Options{User: tc.user})
			if (tc.fail && !errors.Is(err, wantErr)) || (!tc.fail && err != nil) || called != tc.seed {
				t.Fatalf("%v called=%v", err, called)
			}
			var count int64
			if err := db.Model(&goauth.User{}).Where("username = ?", "user").Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if (tc.user == nil && count != 0) || (tc.user != nil && count != 1) {
				t.Fatalf("optional user count=%d", count)
			}
			pool, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			if err := pool.Ping(); err != nil {
				t.Fatalf("caller pool closed: %v", err)
			}
		})
	}
}
