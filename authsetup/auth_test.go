package authsetup

import (
	"testing"

	"github.com/cyphix/goapp/config"
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
		if err != nil {
			t.Fatal(err)
		}
		if got.InitAdminUsername != cfg.Auth.InitAdminUsername || got.InitAdminPassword != cfg.Auth.InitAdminPasswd || got.RPID != cfg.Auth.RPID || got.DevAutoLogin != (env == "development") {
			t.Fatalf("%s: unexpected auth configuration", env)
		}
	}
	cfg.Server.TrustedProxies = "invalid"
	if _, err := AuthConfig(cfg, nil, nil); err == nil {
		t.Fatal("invalid proxy accepted")
	}
}
