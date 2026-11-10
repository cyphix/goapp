package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseValidation(t *testing.T) {
	for _, name := range []string{"", "public", "custom_1", "_private"} {
		db := Database{Schema: name}
		require.Empty(t, db.Validate(), "%q", name)
		require.NotEmpty(t, db.Schema)
	}
	for _, name := range []string{"Upper", "a-b", "1bad", "a;DROP SCHEMA public"} {
		db := Database{Schema: name}
		require.NotEmpty(t, db.Validate(), "%q accepted", name)
	}
	db := Database{Type: "sqlite"}
	require.NotEmpty(t, db.Validate(), "sqlite accepted")
}
