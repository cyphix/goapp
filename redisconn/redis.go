package redisconn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/cyphix/goapp/config"
	"github.com/redis/go-redis/v9"
)

func OpenRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	host, port := cfg.Host, cfg.Port
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port), Password: cfg.Password, DB: cfg.DB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	return rdb, nil
}
