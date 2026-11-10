package dbquery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cyphix/goapp/dbquery"
	logg "github.com/cyphix/gologg"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestQuietNotFound(t *testing.T) {
	previousLevel, previousTimeFormat := zerolog.GlobalLevel(), zerolog.TimeFieldFormat
	t.Cleanup(func() {
		logg.InitWithWriter(previousLevel.String(), os.Stderr)
		zerolog.TimeFieldFormat = previousTimeFormat
	})
	var debugOutput, gormOutput bytes.Buffer
	logg.InitWithWriter("debug", &debugOutput)
	parentLogger := logger.New(log.New(&gormOutput, "", 0), logger.Config{LogLevel: logger.Info})
	db := &gorm.DB{Config: &gorm.Config{Logger: parentLogger}}
	scoped := dbquery.QuietNotFound(db)
	if db.Logger != parentLogger {
		t.Fatal("scope changed the parent logger")
	}
	sql := func() (string, int64) { return "SELECT * FROM user_avatars WHERE user_id = 1", 0 }
	begin := time.Now().Add(-time.Millisecond)
	err := gorm.ErrRecordNotFound
	_, queryFile, queryLine, _ := runtime.Caller(0)
	scoped.Logger.Trace(context.Background(), begin, sql, err)
	if !errors.Is(err, gorm.ErrRecordNotFound) || gormOutput.Len() != 0 {
		t.Fatalf("missing record: err = %v, GORM output = %q", err, &gormOutput)
	}
	var event struct {
		Level   string  `json:"level"`
		SQL     string  `json:"sql"`
		Rows    int64   `json:"rows"`
		Elapsed float64 `json:"elapsed"`
		Caller  string  `json:"caller"`
		Message string  `json:"message"`
	}
	if err := json.Unmarshal(debugOutput.Bytes(), &event); err != nil {
		t.Fatalf("invalid debug output %q: %v", &debugOutput, err)
	}
	querySQL, _ := sql()
	if event.Level != "debug" || event.SQL != querySQL || event.Rows != 0 ||
		event.Elapsed <= 0 || event.Message != "record not found" {
		t.Fatalf("debug event = %+v", event)
	}
	if want := queryFile + ":" + strconv.Itoa(queryLine+1); event.Caller != want {
		t.Fatalf("caller = %q, want %q", event.Caller, want)
	}

	debugOutput.Reset()
	logg.InitWithWriter("info", &debugOutput)
	scoped.Logger.LogMode(logger.Info).Trace(context.Background(), begin, sql, gorm.ErrRecordNotFound)
	if debugOutput.Len() != 0 || gormOutput.Len() != 0 {
		t.Fatalf("missing query logged at info: gologg = %q, GORM = %q", &debugOutput, &gormOutput)
	}

	scoped.Logger.Trace(context.Background(), begin, sql, nil)
	if !strings.Contains(gormOutput.String(), querySQL) {
		t.Fatalf("successful query was not delegated: %s", &gormOutput)
	}
	gormOutput.Reset()
	scoped.Logger.Trace(context.Background(), begin, sql, errors.New("database unavailable"))
	if !strings.Contains(gormOutput.String(), "database unavailable") {
		t.Fatalf("database error was not delegated: %s", &gormOutput)
	}
	if debugOutput.Len() != 0 {
		t.Fatalf("non-not-found traces were logged through gologg: %s", &debugOutput)
	}
}
