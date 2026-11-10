package authsetup

import (
	"testing"

	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig(t *testing.T) {
	cfg := &config.Shared{}
	cfg.Auth.InitAdminUsername = "admin"
	cfg.Auth.InitAdminPasswd = "test-password"
	cfg.Auth.RPID = "localhost"
	cfg.Auth.DevAutoLogin = true
	for _, env := range []string{"development", "production", "test", ""} {
		cfg.Server.Env = env
		got, err := AuthConfig(cfg, nil, nil)
		require.NoError(t, err)
		require.Equal(t, cfg.Auth.InitAdminUsername, got.InitAdminUsername, env)
		require.Equal(t, cfg.Auth.InitAdminPasswd, got.InitAdminPassword, env)
		require.Equal(t, cfg.Auth.RPID, got.RPID, env)
		require.Equal(t, env == "development", got.DevAutoLogin, env)
	}
	cfg.Server.TrustedProxies = "invalid"
	_, err := AuthConfig(cfg, nil, nil)
	require.Error(t, err, "invalid proxy accepted")
}
