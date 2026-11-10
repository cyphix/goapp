// Package migratecmd implements the shared migration command.
package migratecmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/cyphix/goapp/config"
	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/migrate"
	logg "github.com/cyphix/gologg"
	"gorm.io/gorm"
)

type Configuration interface {
	SharedConfig() *config.Shared
	ConfigWarnings() []string
}

type SeedFunc func(context.Context, *gorm.DB) error

// Main loads configuration, logs warnings and exits with status 1 on failure.
func Main[T Configuration](load func() (T, error), src fs.FS, seed func(context.Context, T, *gorm.DB) error) {
	cfg, err := load()
	if err != nil {
		logg.InitConsole("info")
		logger := logg.Ctx("migrate", "main")
		logger.Fatal().Err(err).Msg("failed to load configuration")
	}
	logg.InitConsole(cfg.SharedConfig().Log.Level)
	logger := logg.Ctx("migrate", "main")
	for _, warning := range cfg.ConfigWarnings() {
		logger.Warn().Msg(warning)
	}
	var hook SeedFunc
	if seed != nil {
		hook = func(ctx context.Context, db *gorm.DB) error { return seed(ctx, cfg, db) }
	}
	if err := Run(cfg.SharedConfig(), src, os.Args[1:], hook); err != nil {
		logger.Fatal().Err(err).Msg("migration command failed")
	}
}

func command(args []string) (string, int, error) {
	usage := errors.New("usage: migrate up | down [N] | status | force <version> | reset")
	if len(args) == 0 {
		return "", 0, usage
	}
	switch args[0] {
	case "up", "status", "reset":
		if len(args) != 1 {
			return "", 0, usage
		}
		return args[0], 0, nil
	case "down", "force":
		if args[0] == "down" && len(args) == 1 {
			return "down", 1, nil
		}
		if len(args) != 2 {
			return "", 0, usage
		}
		n, err := strconv.Atoi(args[1])
		if err != nil || (args[0] == "down" && n <= 0) || (args[0] == "force" && n < -1) {
			return "", 0, usage
		}
		return args[0], n, nil
	default:
		return "", 0, usage
	}
}

// Run executes one command without exiting the process.
func Run(cfg *config.Shared, src fs.FS, args []string, seed SeedFunc) error {
	cmd, n, err := command(args)
	if err != nil {
		return err
	}
	if cfg == nil {
		return errors.New("configuration is required")
	}
	if cfg.Database.Type == "" {
		return errors.New("DATABASE_TYPE is required")
	}
	if cmd == "reset" && cfg.Server.Env != "development" {
		return errors.New("reset requires SERVER_ENV=development")
	}
	if cmd == "reset" && (cfg.Auth.InitAdminUsername == "" || cfg.Auth.InitAdminPasswd == "") {
		return errors.New("reset requires AUTH_INIT_ADMIN_USERNAME and AUTH_INIT_ADMIN_PASSWD")
	}
	if cmd == "reset" && seed == nil {
		return errors.New("reset requires a seed function")
	}
	db, err := database.Open(&cfg.Database)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("getting sql connection: %w", err)
	}
	defer sqlDB.Close()
	logger := logg.Ctx("migrate", "main")
	runner := migrate.New(src, cfg.Database.Schema)
	switch cmd {
	case "up":
		err = runner.Up(sqlDB)
	case "down":
		err = runner.Down(sqlDB, n)
	case "force":
		err = runner.Force(sqlDB, n)
	case "status":
		var version uint
		var dirty bool
		version, dirty, err = runner.Status(sqlDB)
		if err == nil {
			event := logger.Info().Uint("version", version).Bool("dirty", dirty)
			if version == 0 && !dirty {
				event.Msg("no migrations applied")
			} else {
				event.Msg("migration status")
			}
		}
	case "reset":
		if err = runner.Reset(sqlDB); err != nil {
			return err
		}
		err = seed(context.Background(), db)
	}
	if err == nil && cmd != "status" {
		logger.Info().Str("command", cmd).Msg("migration complete")
	}
	return err
}
