package migratecmd

import (
	"testing"

	"github.com/cyphix/goapp/config"
)

func TestCommand(t *testing.T) {
	for _, tc := range []struct {
		args []string
		cmd  string
		n    int
	}{
		{[]string{"up"}, "up", 0},
		{[]string{"status"}, "status", 0},
		{[]string{"reset"}, "reset", 0},
		{[]string{"down"}, "down", 1},
		{[]string{"down", "2"}, "down", 2},
		{[]string{"force", "-1"}, "force", -1},
		{[]string{"force", "1"}, "force", 1},
	} {
		cmd, n, err := command(tc.args)
		if err != nil || cmd != tc.cmd || n != tc.n {
			t.Fatalf("%v = %q, %d, %v", tc.args, cmd, n, err)
		}
	}
	for _, args := range [][]string{nil, {"bogus"}, {"up", "1"}, {"down", "0"}, {"down", "-1"}, {"down", "a"}, {"force"}, {"force", "-2"}, {"status", "extra"}} {
		if _, _, err := command(args); err == nil {
			t.Fatalf("%v accepted", args)
		}
	}
}

func TestRunGuards(t *testing.T) {
	if err := Run(nil, nil, []string{"up"}, nil); err == nil {
		t.Fatal("missing configuration accepted")
	}
	cfg := &config.Shared{}
	if err := Run(cfg, nil, []string{"up"}, nil); err == nil {
		t.Fatal("missing database accepted")
	}
	cfg.Database.Type = "postgres"
	for _, env := range []string{"production", "test", ""} {
		cfg.Server.Env = env
		if err := Run(cfg, nil, []string{"reset"}, nil); err == nil || err.Error() != "reset requires SERVER_ENV=development" {
			t.Fatalf("%s: %v", env, err)
		}
	}
	cfg.Server.Env = "development"
	if err := Run(cfg, nil, []string{"reset"}, nil); err == nil || err.Error() != "reset requires AUTH_INIT_ADMIN_USERNAME and AUTH_INIT_ADMIN_PASSWD" {
		t.Fatalf("missing admin credentials: %v", err)
	}
	cfg.Auth.InitAdminUsername, cfg.Auth.InitAdminPasswd = "admin", "test-password"
	if err := Run(cfg, nil, []string{"reset"}, nil); err == nil || err.Error() != "reset requires a seed function" {
		t.Fatalf("missing seed hook: %v", err)
	}
}
