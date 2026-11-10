package database

import (
	"testing"

	"github.com/cyphix/goapp/config"
)

func TestDSNValue(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"", "''"},
		{"plain", "'plain'"},
		{"space quote' slash\\", "'space quote\\' slash\\\\'"},
	} {
		if got := dsnValue(tc.input); got != tc.want {
			t.Fatalf("%q: %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestOpenRejectsNonPostgres(t *testing.T) {
	for _, kind := range []string{"", "sqlite", "mysql"} {
		cfg := &config.Database{}
		cfg.Type = kind
		if _, err := Open(cfg); err == nil {
			t.Fatalf("%q accepted", kind)
		}
	}
}
