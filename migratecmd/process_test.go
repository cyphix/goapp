package migratecmd

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
)

type processConfig struct {
	config.Shared
}

func (c *processConfig) ConfigWarnings() []string { return []string{"example warning"} }

func TestCLIProcess(t *testing.T) {
	mode := os.Getenv("GOAPP_CLI_TEST")
	if mode == "" {
		return
	}
	os.Args = []string{"migrate", "bogus"}
	Main(func() (*processConfig, error) {
		if mode == "config" {
			return nil, errors.New("bad config")
		}
		cfg := &processConfig{}
		cfg.Log.Level = "info"
		return cfg, nil
	}, fstest.MapFS{}, nil)
}

func TestMainFailureLogging(t *testing.T) {
	for _, tc := range []struct {
		mode     string
		messages []string
	}{
		{"config", []string{"bad config", "failed to load configuration"}},
		{"command", []string{"example warning", "usage: migrate", "migration command failed"}},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestCLIProcess$")
			cmd.Env = append(os.Environ(), "GOAPP_CLI_TEST="+tc.mode)
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			require.ErrorAs(t, err, &exit, "%s", output)
			require.Equal(t, 1, exit.ExitCode(), "%s", output)
			for _, message := range tc.messages {
				require.Contains(t, string(output), message)
			}
		})
	}
}
