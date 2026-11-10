//go:build integration

package testdb_test

import (
	"os"
	"testing"

	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }
func TestOpenAppliesCallerMigrations(t *testing.T) {
	first, second := testdb.Open(t), testdb.Open(t)
	if err := first.Exec("INSERT INTO items (label) VALUES (?)", "isolated").Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := second.Table("items").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("database isolation: %d %v", count, err)
	}
}
