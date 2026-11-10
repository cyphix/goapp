// Package testdb provides PostgreSQL databases for integration tests: the
// server at TEST_DATABASE_URL when set, otherwise one ephemeral
// testcontainers instance per test binary.
package testdb

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cyphix/goapp/config"
	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/migrate"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	baseURL    *url.URL
	admin      *gorm.DB
	skipReason string
	seq        atomic.Int64
	source     fs.FS
)

// Main runs the package's tests against a shared PostgreSQL server and
// terminates any container it started. Call it from TestMain.
func Main(m *testing.M, src fs.FS) int { source = src; return run(m) }

func run(m *testing.M) int {
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		ctr, err := startContainer(ctx, "postgres:17-alpine")
		if ctr != nil {
			defer func() { _ = ctr.Terminate(context.Background()) }()
		}
		if err == nil {
			raw, err = ctr.ConnectionString(ctx, "sslmode=disable")
		}
		cancel()
		if err != nil {
			skipReason = fmt.Sprintf("postgres unavailable: %v", err)
			return m.Run()
		}
	}

	u, err := url.Parse(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testdb: parsing database URL: %v\n", err)
		return 1
	}
	baseURL = u
	admin, err = gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		fmt.Fprintf(os.Stderr, "testdb: connecting: %v\n", err)
		return 1
	}
	sqlDB, err := admin.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testdb: sql connection: %v\n", err)
		return 1
	}
	defer sqlDB.Close()
	return m.Run()
}

// Config creates an empty database that is dropped when t finishes.
func Config(t testing.TB) *config.Shared {
	t.Helper()
	if admin == nil {
		if skipReason == "" {
			skipReason = "testdb.Main was not called from TestMain"
		}
		t.Skip(skipReason)
	}

	name := fmt.Sprintf("test_%d_%d", os.Getpid(), seq.Add(1))
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatalf("testdb: create database: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)").Error; err != nil {
			t.Logf("testdb: drop database: %v", err)
		}
	})

	u := *baseURL
	u.Path = "/" + name

	port := 5432
	if u.Port() != "" {
		var err error
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			t.Fatal(err)
		}
	}
	cfg := &config.Shared{}
	cfg.Database = config.Database{Type: "postgres", Host: u.Hostname(), Port: port, Name: name, SSLMode: u.Query().Get("sslmode"), Schema: "public"}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if u.User != nil {
		cfg.Database.User = u.User.Username()
		cfg.Database.Password, _ = u.User.Password()
	}
	return cfg
}

func Open(t testing.TB) *gorm.DB {
	t.Helper()
	cfg := Config(t)
	db, err := database.Open(&cfg.Database)
	if err != nil {
		t.Fatalf("testdb: open: %v", err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := migrate.New(source, cfg.Database.Schema).Up(pool); err != nil {
		t.Fatalf("testdb: migrate: %v", err)
	}
	return db
}

func startContainer(ctx context.Context, image string) (ctr *tcpostgres.PostgresContainer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("docker unavailable: %v", r)
		}
	}()
	return tcpostgres.Run(ctx, image,
		tcpostgres.WithDatabase("postgres"), tcpostgres.WithUsername("test"), tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)),
	)
}
