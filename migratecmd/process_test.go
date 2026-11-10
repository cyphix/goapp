package migratecmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/config"
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
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("exit: %v, %s", err, output)
			}
			for _, message := range tc.messages {
				if !strings.Contains(string(output), message) {
					t.Fatalf("missing %q in %s", message, output)
				}
			}
		})
	}
}
