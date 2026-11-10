package testdb_test

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/testdb"
)

func TestFailureProcess(t *testing.T) {
	if os.Getenv("GOAPP_TESTDB_FAILURE") == "" {
		return
	}
	os.Exit(testdb.Main(nil, fstest.MapFS{}))
}

func TestExplicitURLFailure(t *testing.T) {
	for _, tc := range []struct{ url, message string }{
		{"%", "testdb: parsing database URL:"},
		{"postgres://test:test-password@127.0.0.1:1/postgres?sslmode=disable&connect_timeout=1", "testdb: connecting:"},
	} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestFailureProcess$")
		cmd.Env = append(os.Environ(), "GOAPP_TESTDB_FAILURE=1", "TEST_DATABASE_URL="+tc.url)
		output, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), tc.message) {
			t.Fatalf("exit: %v, output: %s", err, output)
		}
	}
}
