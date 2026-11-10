// Package devseed bootstraps auth and invokes app-owned development data hooks.
package devseed

import (
	"context"
	"fmt"
	"github.com/cyphix/goapp/authsetup"
	"github.com/cyphix/goapp/config"
	"github.com/cyphix/goapp/redisconn"
	"github.com/cyphix/goauth"
	logg "github.com/cyphix/gologg"
	"gorm.io/gorm"
)

type User struct{ Username, Password string }
type Options struct{ User *User }
type SeedFunc func(context.Context, *gorm.DB) error

// Run owns its Redis connection, but leaves the caller's database open.
// A nil User skips the optional development account; a nil seed skips domain data.
func Run(ctx context.Context, cfg *config.Shared, db *gorm.DB, seed SeedFunc, options Options) error {
	if options.User != nil && (options.User.Username == "" || options.User.Password == "") {
		return fmt.Errorf("seed auth user: username and password are required")
	}
	rdb, err := redisconn.OpenRedis(&cfg.Redis)
	if err != nil {
		return err
	}
	defer rdb.Close()
	auth, err := authsetup.NewAuth(cfg, db, rdb)
	if err != nil {
		return fmt.Errorf("bootstrapping auth: %w", err)
	}
	if user := options.User; user != nil {
		if _, err := auth.EnsureUser(user.Username, user.Password, goauth.RoleUser); err != nil {
			return fmt.Errorf("seed auth user: %w", err)
		}
		logger := logg.Ctx("devseed", "auth")
		logger.Info().Str("username", user.Username).Msg("dev auth user ensured")
	}
	if seed == nil {
		return nil
	}
	return seed(ctx, db)
}
