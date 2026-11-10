package migratecmd

import (
	"testing"

	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, err, "%v", tc.args)
		require.Equal(t, tc.cmd, cmd, "%v", tc.args)
		require.Equal(t, tc.n, n, "%v", tc.args)
	}
	for _, args := range [][]string{nil, {"bogus"}, {"up", "1"}, {"down", "0"}, {"down", "-1"}, {"down", "a"}, {"force"}, {"force", "-2"}, {"status", "extra"}} {
		_, _, err := command(args)
		require.Error(t, err, "%v accepted", args)
	}
}

func TestRunGuards(t *testing.T) {
	require.Error(t, Run(nil, nil, []string{"up"}, nil), "missing configuration accepted")
	cfg := &config.Shared{}
	require.Error(t, Run(cfg, nil, []string{"up"}, nil), "missing database accepted")
	cfg.Database.Type = "postgres"
	for _, env := range []string{"production", "test", ""} {
		cfg.Server.Env = env
		require.EqualError(t, Run(cfg, nil, []string{"reset"}, nil), "reset requires SERVER_ENV=development", env)
	}
	cfg.Server.Env = "development"
	require.EqualError(t, Run(cfg, nil, []string{"reset"}, nil), "reset requires AUTH_INIT_ADMIN_USERNAME and AUTH_INIT_ADMIN_PASSWD")
	cfg.Auth.InitAdminUsername, cfg.Auth.InitAdminPasswd = "admin", "test-password"
	require.EqualError(t, Run(cfg, nil, []string{"reset"}, nil), "reset requires a seed function")
}
