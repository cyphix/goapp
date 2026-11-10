package redisconn

import (
	"context"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
	"github.com/stretchr/testify/require"
)

func TestOpenRedis(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("test-password")
	_, port, err := net.SplitHostPort(server.Addr())
	require.NoError(t, err)
	cfg := &config.RedisConfig{Port: port, Password: "test-password", DB: 2}
	client, err := OpenRedis(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.Equal(t, net.JoinHostPort("localhost", port), client.Options().Addr)
	require.NoError(t, client.Set(context.Background(), "example", "value", 0).Err())
	got, err := server.DB(2).Get("example")
	require.NoError(t, err)
	require.Equal(t, "value", got)
}
func TestOpenRedisError(t *testing.T) {
	client, err := OpenRedis(&config.RedisConfig{Host: "127.0.0.1", Port: "invalid"})
	require.Nil(t, client)
	require.ErrorContains(t, err, "connecting to redis:")
}
