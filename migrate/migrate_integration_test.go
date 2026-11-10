//go:build integration

package migrate_test

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/cyphix/goapp/database"
	"github.com/cyphix/goapp/internal/testfixtures"
	"github.com/cyphix/goapp/migrate"
	"github.com/cyphix/goapp/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m, testfixtures.FS)) }

func TestLifecycle(t *testing.T) {
	cfg := testdb.Config(t)
	db, err := database.Open(&cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	status := func(want uint) {
		t.Helper()
		version, dirty, err := migrate.New(testfixtures.FS, "").Status(sqlDB)
		if err != nil || version != want || dirty {
			t.Fatalf("status = (%d, %v, %v), want (%d, false, nil)", version, dirty, err, want)
		}
		if err := sqlDB.Ping(); err != nil {
			t.Fatalf("caller pool closed: %v", err)
		}
	}
	status(0)
	if err := migrate.New(testfixtures.FS, "").Up(sqlDB); err != nil {
		t.Fatal(err)
	}
	status(1)
	if err := migrate.New(testfixtures.FS, "").Down(sqlDB, 1); err != nil {
		t.Fatal(err)
	}
	status(0)
	if err := migrate.New(testfixtures.FS, "").Up(sqlDB); err != nil {
		t.Fatal(err)
	}
	status(1)
	if _, err := sqlDB.Exec("UPDATE public.schema_migrations SET dirty = true"); err != nil {
		t.Fatal(err)
	}
	if version, dirty, err := migrate.New(testfixtures.FS, "").Status(sqlDB); err != nil || version != 1 || !dirty {
		t.Fatalf("dirty status = (%d, %v, %v)", version, dirty, err)
	}
	if err := migrate.New(testfixtures.FS, "").Up(sqlDB); err == nil {
		t.Fatal("dirty database accepted")
	}
	if err := migrate.New(testfixtures.FS, "").Force(sqlDB, 1); err != nil {
		t.Fatal(err)
	}
	status(1)
	if err := migrate.New(testfixtures.FS, "").Down(sqlDB, 0); err == nil {
		t.Fatal("zero steps accepted")
	}
	if err := migrate.New(testfixtures.FS, "").Reset(sqlDB); err != nil {
		t.Fatal(err)
	}
	status(1)
	if err := migrate.New(testfixtures.FS, "public").Reset(sqlDB); err != nil {
		t.Fatal(err)
	}
	status(1)
	if err := migrate.New(testfixtures.FS, "").Force(sqlDB, -2); err == nil {
		t.Fatal("invalid version accepted")
	}
	if _, err := sqlDB.Exec(`CREATE SCHEMA "quoted""schema"`); err != nil {
		t.Fatal(err)
	}
	if err := migrate.New(testfixtures.FS, `quoted"schema`).Reset(sqlDB); err != nil {
		t.Fatal(err)
	}
	status(1)
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := migrate.New(testfixtures.FS, "").Up(sqlDB); err == nil {
		t.Fatal("closed pool accepted")
	}
}

func TestCustomSchema(t *testing.T) {
	db := testdb.Open(t)
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	for _, name := range []string{"custom_schema", `quoted"schema`} {
		if err := migrate.New(testfixtures.FS, name).Up(pool); err != nil {
			t.Fatal(err)
		}
		version, dirty, err := migrate.New(testfixtures.FS, name).Status(pool)
		if err != nil || dirty || version == 0 {
			t.Fatalf("status: %d %v %v", version, dirty, err)
		}
		if err := migrate.New(testfixtures.FS, name).Reset(pool); err != nil {
			t.Fatal(err)
		}
	}
	var path string
	if err := pool.QueryRow("SHOW search_path").Scan(&path); err != nil {
		t.Fatal(err)
	}
	if path != `"public"` {
		t.Fatalf("search path leaked: %q", path)
	}
}

func TestFilesystemFailures(t *testing.T) {
	db := testdb.Open(t)
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	for _, runner := range []*migrate.Runner{
		migrate.New(nil, ""),
		migrate.New(os.DirFS(t.TempDir()+"/missing"), ""),
		migrate.New(fstest.MapFS{"000001_bad.up.sql": &fstest.MapFile{Data: []byte("INVALID SQL;")}}, "broken"),
	} {
		if err := runner.Up(pool); err == nil {
			t.Fatal("invalid filesystem/migration accepted")
		}
		if err := pool.Ping(); err != nil {
			t.Fatalf("caller pool closed: %v", err)
		}
	}
	empty := migrate.New(fstest.MapFS{}, "empty")
	version, dirty, err := empty.Status(pool)
	if err != nil || version != 0 || dirty {
		t.Fatalf("empty source status: %d %v %v", version, dirty, err)
	}
}
