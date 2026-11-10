package devseed

import (
	"context"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRunRejectsIncompleteUser(t *testing.T) {
	for _, user := range []*User{{}, {Username: "user"}, {Password: "test-password"}} {
		require.Error(t, Run(context.Background(), &config.Shared{}, nil, nil, Options{User: user}), "incomplete user accepted")
	}
}
func TestRunAuthError(t *testing.T) {
	server := miniredis.RunT(t)
	host, port, err := net.SplitHostPort(server.Addr())
	require.NoError(t, err)
	cfg := &config.Shared{}
	cfg.Redis = config.RedisConfig{Host: host, Port: port}
	cfg.Server.TrustedProxies = "invalid"
	called := false
	err = Run(context.Background(), cfg, nil, func(context.Context, *gorm.DB) error { called = true; return nil }, Options{})
	require.ErrorContains(t, err, "bootstrapping auth: invalid SERVER_TRUSTED_PROXIES")
	require.False(t, called)
}
