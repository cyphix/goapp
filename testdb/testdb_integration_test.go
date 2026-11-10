//go:build integration

package testdb_test

import (
	"os"
	"testing"

	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestOpenAppliesCallerMigrations(t *testing.T) {
	first, second := testdb.Open(t), testdb.Open(t)
	require.NoError(t, first.Exec("INSERT INTO items (label) VALUES (?)", "isolated").Error)
	var count int64
	require.NoError(t, second.Table("items").Count(&count).Error)
	require.Zero(t, count)
}
