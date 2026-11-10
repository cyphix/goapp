package config

import (
	"os"
	"reflect"
	"strings"
	"testing"
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
	if err != nil {
		t.Fatal(err)
	}
	want := Shared{
		Server:   ServerConfig{Env: "development", Port: 1323},
		Log:      LogConfig{Level: "info"},
		Database: Database{Host: "localhost", Port: 5432, SSLMode: "disable", Schema: "public"},
		Redis:    RedisConfig{Host: "localhost", Port: "6379"},
		Auth:     AuthConfig{RPID: "localhost", RPDisplayName: newConfig().Auth.RPDisplayName, Origins: []string{"http://localhost:5173"}, PublicBaseURL: "http://localhost:5173"},
	}
	if !reflect.DeepEqual(cfg.Shared, want) {
		t.Fatalf("shared defaults = %#v, want %#v", cfg.Shared, want)
	}
	if len(cfg.Warnings) != 0 {
		t.Fatalf("warnings = %v", cfg.Warnings)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	want := Shared{
		Server:   ServerConfig{Env: "production", Port: 9090, TrustedProxies: "10.0.0.0/8"},
		Log:      LogConfig{Level: "debug"},
		Database: Database{Type: "postgres", Host: "db.example", Port: 5433, User: "example", Password: "example-password", Name: "example_db", SSLMode: "require", Schema: "example_schema"},
		Redis:    RedisConfig{Host: "cache.example", Port: "6380", Password: "example-cache-password", DB: 2},
		Auth:     AuthConfig{InitAdminUsername: "example_admin", InitAdminPasswd: "example-admin-password", RPID: "example.com", RPDisplayName: "Example App", Origins: []string{"https://example.com", "https://other.example.com"}, PublicBaseURL: "https://example.com", CookieSecure: true, DevAutoLogin: true},
	}
	if !reflect.DeepEqual(cfg.Shared, want) {
		t.Fatalf("shared environment mapping differs")
	}
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
			if err == nil || cfg != nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestSharedDotEnvAndWarnings(t *testing.T) {
	t.Run("file and environment precedence", func(t *testing.T) {
		resetShared(t)
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(".env", []byte("LOG_LEVEL=debug\nSERVER_PORT=9999\n"), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("SERVER_PORT", "8888")
		cfg, err := loadFixture()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Log.Level != "debug" || cfg.Server.Port != 8888 {
			t.Fatal("dotenv mapping or precedence differs")
		}
	})
	t.Run("malformed optional file", func(t *testing.T) {
		resetShared(t)
		if err := os.WriteFile(".env", []byte("invalid content !"), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := loadFixture()
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], "could not load .env") {
			t.Fatalf("warnings=%v", cfg.Warnings)
		}
	})
}

func TestSharedLegacyNamesAreNotAliases(t *testing.T) {
	resetShared(t)
	t.Setenv("AUTH_RPID", "legacy.example")
	t.Setenv("DEV_AUTO_LOGIN", "true")
	cfg, err := loadFixture()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.RPID != "localhost" || cfg.Auth.DevAutoLogin {
		t.Fatal("obsolete auth environment names were accepted")
	}
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
		if c.Label != "custom" {
			t.Fatalf("app environment not parsed: %q", c.Label)
		}
		return nil
	})
	if err != nil || len(warnings) != 0 || !called {
		t.Fatalf("%v %v %v", warnings, err, called)
	}
	t.Setenv("SERVER_PORT", "invalid")
	t.Setenv("SERVER_ENV", "staging")
	_, err = Load(cfg, func(c *fixture) *Shared { return &c.Shared }, func(c *fixture) []string { return []string{"APP_LABEL rejected"} })
	if err == nil {
		t.Fatal("invalid app config accepted")
	}
	for _, want := range []string{"environment parse error:", "SERVER_ENV", "APP_LABEL rejected"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("missing %q in %v", want, err)
		}
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
		if _, err := Load(tc.cfg, tc.shared, nil); err == nil {
			t.Fatal("missing configuration accepted")
		}
	}
}
