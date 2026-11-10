package database

import (
	"testing"

	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
)

func TestDSNValue(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"", "''"},
		{"plain", "'plain'"},
		{"space quote' slash\\", "'space quote\\' slash\\\\'"},
	} {
		require.Equal(t, tc.want, dsnValue(tc.input), "%q", tc.input)
	}
}

func TestOpenRejectsNonPostgres(t *testing.T) {
	for _, kind := range []string{"", "sqlite", "mysql"} {
		cfg := &config.Database{}
		cfg.Type = kind
		_, err := Open(cfg)
		require.Error(t, err, "%q accepted", kind)
	}
}
