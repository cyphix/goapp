//go:build integration

package devseed_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
	"github.com/cyphix/goapp/devseed"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
	"github.com/cyphix/goauth"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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
			require.NoError(t, err)
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
					require.Same(t, ctx, got)
					require.Same(t, db, gotDB)
					var count int64
					require.NoError(t, db.Model(&goauth.User{}).Where("username = ?", "admin").Count(&count).Error)
					require.Equal(t, int64(1), count)
					if tc.fail {
						return wantErr
					}
					return db.Exec("INSERT INTO items (label) VALUES (?)", "seeded").Error
				}
			}
			err = devseed.Run(ctx, cfg, db, hook, devseed.Options{User: tc.user})
			if tc.fail {
				require.ErrorIs(t, err, wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.seed, called)
			var count int64
			require.NoError(t, db.Model(&goauth.User{}).Where("username = ?", "user").Count(&count).Error)
			if tc.user == nil {
				require.Zero(t, count)
			} else {
				require.Equal(t, int64(1), count)
			}
			pool, err := db.DB()
			require.NoError(t, err)
			require.NoError(t, pool.Ping(), "caller pool closed")
		})
	}
}
