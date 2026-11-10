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
	"testing"
	"time"

	"github.com/cyphix/goapp/dbquery"
	logg "github.com/cyphix/gologg"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
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
	require.Same(t, parentLogger, db.Logger)
	sql := func() (string, int64) { return "SELECT * FROM user_avatars WHERE user_id = 1", 0 }
	begin := time.Now().Add(-time.Millisecond)
	err := gorm.ErrRecordNotFound
	_, queryFile, queryLine, _ := runtime.Caller(0)
	scoped.Logger.Trace(context.Background(), begin, sql, err)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Empty(t, gormOutput.String())
	var event struct {
		Level   string  `json:"level"`
		SQL     string  `json:"sql"`
		Rows    int64   `json:"rows"`
		Elapsed float64 `json:"elapsed"`
		Caller  string  `json:"caller"`
		Message string  `json:"message"`
	}
	require.NoError(t, json.Unmarshal(debugOutput.Bytes(), &event), "%s", &debugOutput)
	querySQL, _ := sql()
	require.Equal(t, "debug", event.Level)
	require.Equal(t, querySQL, event.SQL)
	require.Zero(t, event.Rows)
	require.Positive(t, event.Elapsed)
	require.Equal(t, "record not found", event.Message)
	require.Equal(t, queryFile+":"+strconv.Itoa(queryLine+1), event.Caller)

	debugOutput.Reset()
	logg.InitWithWriter("info", &debugOutput)
	scoped.Logger.LogMode(logger.Info).Trace(context.Background(), begin, sql, gorm.ErrRecordNotFound)
	require.Empty(t, debugOutput.String())
	require.Empty(t, gormOutput.String())

	scoped.Logger.Trace(context.Background(), begin, sql, nil)
	require.Contains(t, gormOutput.String(), querySQL)
	gormOutput.Reset()
	scoped.Logger.Trace(context.Background(), begin, sql, errors.New("database unavailable"))
	require.Contains(t, gormOutput.String(), "database unavailable")
	require.Empty(t, debugOutput.String())
}
