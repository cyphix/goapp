package config

import (
	"fmt"
	"strings"
)

type Shared struct {
	Server   ServerConfig `envPrefix:"SERVER_"`
	Log      LogConfig    `envPrefix:"LOG_"`
	Database Database     `envPrefix:"DATABASE_"`
	Redis    RedisConfig  `envPrefix:"REDIS_"`
	Auth     AuthConfig   `envPrefix:"AUTH_"`
}

type ServerConfig struct {
	Env            string `env:"ENV" envDefault:"development"`
	Port           int    `env:"PORT" envDefault:"1323"`
	TrustedProxies string `env:"TRUSTED_PROXIES"`
}

type LogConfig struct {
	Level string `env:"LEVEL" envDefault:"info"`
}

type RedisConfig struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`
}

type AuthConfig struct {
	InitAdminUsername string   `env:"INIT_ADMIN_USERNAME"`
	InitAdminPasswd   string   `env:"INIT_ADMIN_PASSWD"`
	RPID              string   `env:"RP_ID" envDefault:"localhost"`
	RPDisplayName     string   `env:"RP_DISPLAY_NAME"`
	Origins           []string `env:"ORIGINS" envSeparator:"," envDefault:"http://localhost:5173"`
	PublicBaseURL     string   `env:"PUBLIC_BASE_URL" envDefault:"http://localhost:5173"`
	CookieSecure      bool     `env:"COOKIE_SECURE" envDefault:"false"`
	DevAutoLogin      bool     `env:"DEV_AUTO_LOGIN" envDefault:"false"`
}

func (cfg *Shared) Validate() []string {
	var problems []string
	switch cfg.Server.Env {
	case "development", "test", "production":
	default:
		problems = append(problems, fmt.Sprintf("invalid server env: SERVER_ENV must be one of development|test|production (got %q)", cfg.Server.Env))
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		problems = append(problems, fmt.Sprintf("SERVER_PORT must be between 1 and 65535 (got %d)", cfg.Server.Port))
	}
	switch cfg.Log.Level {
	case "trace", "debug", "info", "warn", "error", "fatal":
	default:
		problems = append(problems, fmt.Sprintf("invalid log level: LOG_LEVEL must be one of trace|debug|info|warn|error|fatal (got %q)", cfg.Log.Level))
	}
	problems = append(problems, cfg.Database.Validate()...)
	for i := range cfg.Auth.Origins {
		cfg.Auth.Origins[i] = strings.TrimSpace(cfg.Auth.Origins[i])
	}
	return problems
}

func (cfg *Shared) SharedConfig() *Shared { return cfg }
