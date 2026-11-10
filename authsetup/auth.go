package authsetup

import (
	"fmt"
	"github.com/cyphix/goapp/httpx"

	"github.com/cyphix/goapp/config"

	"github.com/cyphix/goauth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuthConfig builds the shared server and migration bootstrap auth configuration.
func AuthConfig(cfg *config.Shared, db *gorm.DB, redisDb *redis.Client) (goauth.Config, error) {
	clientIP, err := httpx.ClientIPExtractor(cfg.Server.TrustedProxies)
	if err != nil {
		return goauth.Config{}, fmt.Errorf("invalid SERVER_TRUSTED_PROXIES: %w", err)
	}
	return goauth.Config{
		DB:                db,
		Redis:             redisDb,
		RPDisplayName:     cfg.Auth.RPDisplayName,
		RPID:              cfg.Auth.RPID,
		Origins:           cfg.Auth.Origins,
		PublicBaseURL:     cfg.Auth.PublicBaseURL,
		CookieSecure:      cfg.Auth.CookieSecure,
		ClientIP:          clientIP,
		InitAdminUsername: cfg.Auth.InitAdminUsername,
		InitAdminPassword: cfg.Auth.InitAdminPasswd,
		DevAutoLogin:      cfg.Server.Env == "development" && cfg.Auth.DevAutoLogin,
	}, nil
}

func NewAuth(cfg *config.Shared, db *gorm.DB, rdb *redis.Client) (*goauth.Auth, error) {
	authCfg, err := AuthConfig(cfg, db, rdb)
	if err != nil {
		return nil, err
	}
	return goauth.New(authCfg)
}
