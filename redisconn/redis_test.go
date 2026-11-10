package redisconn

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cyphix/goapp/config"
)

func TestOpenRedis(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("test-password")
	_, port, err := net.SplitHostPort(server.Addr())
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.RedisConfig{Port: port, Password: "test-password", DB: 2}
	client, err := OpenRedis(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if client.Options().Addr != net.JoinHostPort("localhost", port) {
		t.Fatal(client.Options().Addr)
	}
	if err := client.Set(context.Background(), "example", "value", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if got, err := server.DB(2).Get("example"); err != nil || got != "value" {
		t.Fatalf("%q %v", got, err)
	}
}
func TestOpenRedisError(t *testing.T) {
	client, err := OpenRedis(&config.RedisConfig{Host: "127.0.0.1", Port: "invalid"})
	if client != nil || err == nil || !strings.Contains(err.Error(), "connecting to redis:") {
		t.Fatalf("%v %v", client, err)
	}
}
