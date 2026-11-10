package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetShared(t *testing.T) {
	t.Helper()
	t.Setenv("APP_LABEL", "")
	for _, key := range []string{"SERVER_ENV", "SERVER_PORT", "SERVER_TRUSTED_PROXIES", "LOG_LEVEL", "DATABASE_TYPE", "DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER", "DATABASE_PASSWORD", "DATABASE_NAME", "DATABASE_SSLMODE", "DATABASE_SCHEMA", "REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB", "AUTH_INIT_ADMIN_USERNAME", "AUTH_INIT_ADMIN_PASSWD", "AUTH_RP_ID", "AUTH_RP_DISPLAY_NAME", "AUTH_ORIGINS", "AUTH_PUBLIC_BASE_URL", "AUTH_COOKIE_SECURE", "AUTH_DEV_AUTO_LOGIN"} {
		t.Setenv(key, "")
	}
	t.Setenv("AUTH_RPID", "")
	t.Setenv("DEV_AUTO_LOGIN", "")
	t.Chdir(t.TempDir())
}

func TestSharedDefaults(t *testing.T) {
	resetShared(t)
	cfg, err := loadFixture()
	require.NoError(t, err)
	want := Shared{
		Server:   ServerConfig{Env: "development", Port: 1323},
		Log:      LogConfig{Level: "info"},
		Database: Database{Host: "localhost", Port: 5432, SSLMode: "disable", Schema: "public"},
		Redis:    RedisConfig{Host: "localhost", Port: "6379"},
		Auth:     AuthConfig{RPID: "localhost", RPDisplayName: newConfig().Auth.RPDisplayName, Origins: []string{"http://localhost:5173"}, PublicBaseURL: "http://localhost:5173"},
	}
	require.Equal(t, want, cfg.Shared)
	require.Empty(t, cfg.Warnings)
}

func TestSharedEnvironmentMapping(t *testing.T) {
	resetShared(t)
	values := map[string]string{
		"SERVER_ENV": "production", "SERVER_PORT": "9090", "SERVER_TRUSTED_PROXIES": "10.0.0.0/8",
		"LOG_LEVEL":     "debug",
		"DATABASE_TYPE": "postgres", "DATABASE_HOST": "db.example", "DATABASE_PORT": "5433",
		"DATABASE_USER": "example", "DATABASE_PASSWORD": "example-password", "DATABASE_NAME": "example_db", "DATABASE_SSLMODE": "require", "DATABASE_SCHEMA": "example_schema",
		"REDIS_HOST": "cache.example", "REDIS_PORT": "6380", "REDIS_PASSWORD": "example-cache-password", "REDIS_DB": "2",
		"AUTH_INIT_ADMIN_USERNAME": "example_admin", "AUTH_INIT_ADMIN_PASSWD": "example-admin-password", "AUTH_RP_ID": "example.com", "AUTH_RP_DISPLAY_NAME": "Example App",
		"AUTH_ORIGINS": "https://example.com, https://other.example.com", "AUTH_PUBLIC_BASE_URL": "https://example.com", "AUTH_COOKIE_SECURE": "true", "AUTH_DEV_AUTO_LOGIN": "true",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
	cfg, err := loadFixture()
	require.NoError(t, err)
	want := Shared{
		Server:   ServerConfig{Env: "production", Port: 9090, TrustedProxies: "10.0.0.0/8"},
		Log:      LogConfig{Level: "debug"},
		Database: Database{Type: "postgres", Host: "db.example", Port: 5433, User: "example", Password: "example-password", Name: "example_db", SSLMode: "require", Schema: "example_schema"},
		Redis:    RedisConfig{Host: "cache.example", Port: "6380", Password: "example-cache-password", DB: 2},
		Auth:     AuthConfig{InitAdminUsername: "example_admin", InitAdminPasswd: "example-admin-password", RPID: "example.com", RPDisplayName: "Example App", Origins: []string{"https://example.com", "https://other.example.com"}, PublicBaseURL: "https://example.com", CookieSecure: true, DevAutoLogin: true},
	}
	require.Equal(t, want, cfg.Shared)
}

func TestSharedRejectInvalid(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"SERVER_ENV", "dev"}, {"SERVER_ENV", "prod"}, {"SERVER_ENV", "staging"},
		{"SERVER_PORT", "0"}, {"SERVER_PORT", "65536"}, {"SERVER_PORT", "invalid"},
		{"LOG_LEVEL", "invalid"}, {"DATABASE_TYPE", "sqlite"}, {"DATABASE_SCHEMA", "bad-schema"},
		{"DATABASE_PORT", "invalid"}, {"REDIS_DB", "invalid"}, {"AUTH_COOKIE_SECURE", "invalid"},
	} {
		t.Run(tc.key+tc.value, func(t *testing.T) {
			resetShared(t)
			t.Setenv(tc.key, tc.value)
			cfg, err := loadFixture()
			require.Error(t, err)
			require.Nil(t, cfg)
		})
	}
}

func TestSharedDotEnvAndWarnings(t *testing.T) {
	t.Run("file and environment precedence", func(t *testing.T) {
		resetShared(t)
		require.NoError(t, os.Unsetenv("LOG_LEVEL"))
		require.NoError(t, os.WriteFile(".env", []byte("LOG_LEVEL=debug\nSERVER_PORT=9999\n"), 0600))
		t.Setenv("SERVER_PORT", "8888")
		cfg, err := loadFixture()
		require.NoError(t, err)
		require.Equal(t, "debug", cfg.Log.Level)
		require.Equal(t, 8888, cfg.Server.Port)
	})
	t.Run("malformed optional file", func(t *testing.T) {
		resetShared(t)
		require.NoError(t, os.WriteFile(".env", []byte("invalid content !"), 0600))
		cfg, err := loadFixture()
		require.NoError(t, err)
		require.Len(t, cfg.Warnings, 1)
		require.Contains(t, cfg.Warnings[0], "could not load .env")
	})
}

func TestSharedLegacyNamesAreNotAliases(t *testing.T) {
	resetShared(t)
	t.Setenv("AUTH_RPID", "legacy.example")
	t.Setenv("DEV_AUTO_LOGIN", "true")
	cfg, err := loadFixture()
	require.NoError(t, err)
	require.Equal(t, "localhost", cfg.Auth.RPID)
	require.False(t, cfg.Auth.DevAutoLogin)
}

type fixture struct {
	Shared
	Warnings []string `env:"-"`
	Label    string   `env:"APP_LABEL" envDefault:"example"`
}

func newConfig() *fixture {
	cfg := &fixture{}
	cfg.Auth.RPDisplayName = "Example"
	return cfg
}
func loadFixture() (*fixture, error) {
	cfg := newConfig()
	warnings, err := Load(cfg, func(c *fixture) *Shared { return &c.Shared }, nil)
	cfg.Warnings = warnings
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
func TestAppValidationAndParseErrors(t *testing.T) {
	resetShared(t)
	t.Setenv("APP_LABEL", "custom")
	cfg := newConfig()
	var called bool
	warnings, err := Load(cfg, func(c *fixture) *Shared { return &c.Shared }, func(c *fixture) []string {
		called = true
		require.Equal(t, "custom", c.Label)
		return nil
	})
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.True(t, called)
	t.Setenv("SERVER_PORT", "invalid")
	t.Setenv("SERVER_ENV", "staging")
	_, err = Load(cfg, func(c *fixture) *Shared { return &c.Shared }, func(c *fixture) []string { return []string{"APP_LABEL rejected"} })
	require.Error(t, err)
	for _, want := range []string{"environment parse error:", "SERVER_ENV", "APP_LABEL rejected"} {
		require.ErrorContains(t, err, want)
	}
}
func TestLoadMissingInputs(t *testing.T) {
	for _, tc := range []struct {
		cfg    *fixture
		shared func(*fixture) *Shared
	}{
		{nil, func(c *fixture) *Shared { return nil }},
		{newConfig(), nil},
		{newConfig(), func(c *fixture) *Shared { return nil }},
	} {
		_, err := Load(tc.cfg, tc.shared, nil)
		require.Error(t, err, "missing configuration accepted")
	}
}
