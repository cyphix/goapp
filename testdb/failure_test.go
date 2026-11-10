package testdb_test

import (
	"os"
	"os/exec"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/testdb"
	"github.com/stretchr/testify/require"
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
		require.ErrorAs(t, err, &exit, "%s", output)
		require.Equal(t, 1, exit.ExitCode(), "%s", output)
		require.Contains(t, string(output), tc.message)
	}
}
