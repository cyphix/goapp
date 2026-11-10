package devseed

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
	"gorm.io/gorm"
)

func TestRunRejectsIncompleteUser(t *testing.T) {
	for _, user := range []*User{{}, {Username: "user"}, {Password: "test-password"}} {
		if err := Run(context.Background(), &config.Shared{}, nil, nil, Options{User: user}); err == nil {
			t.Fatal("incomplete user accepted")
		}
	}
}
func TestRunAuthError(t *testing.T) {
	server := miniredis.RunT(t)
	host, port, err := net.SplitHostPort(server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Shared{}
	cfg.Redis = config.RedisConfig{Host: host, Port: port}
	cfg.Server.TrustedProxies = "invalid"
	called := false
	err = Run(context.Background(), cfg, nil, func(context.Context, *gorm.DB) error { called = true; return nil }, Options{})
	if err == nil || !strings.Contains(err.Error(), "bootstrapping auth: invalid SERVER_TRUSTED_PROXIES") || called {
		t.Fatalf("%v, hook called=%v", err, called)
	}
}
