package dbquery

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"time"

	logg "github.com/cyphix/gologg"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var queryLogger = logg.Ctx("dbquery", "quiet_not_found")

// QuietNotFound is a query scope that logs not-found errors only at debug level.
// Use it only where a missing row is expected and handled, such as an optional
// avatar or preferences that fall back to defaults. Not-found is usually a useful
// signal; do not apply this scope broadly.
func QuietNotFound(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{Logger: quietNotFoundLogger{db.Logger}})
}

type quietNotFoundLogger struct {
	logger.Interface
}

func (l quietNotFoundLogger) LogMode(level logger.LogLevel) logger.Interface {
	return quietNotFoundLogger{l.Interface.LogMode(level)}
}

func (l quietNotFoundLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Interface.Trace(ctx, begin, fc, err)
		return
	}
	if event := queryLogger.Debug(); event.Enabled() {
		sql, rows := fc()
		event.Str("sql", sql).Int64("rows", rows).
			Dur("elapsed", time.Since(begin)).Str("caller", queryCaller()).
			Msg("record not found")
	}
}

func queryCaller() string {
	var pcs [64]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if frame.Function != "" &&
			!strings.HasPrefix(frame.Function, "gorm.io/gorm.") &&
			!strings.HasPrefix(frame.Function, "gorm.io/gorm/") &&
			!strings.HasPrefix(frame.Function, "github.com/cyphix/goapp/dbquery.") {
			return frame.File + ":" + strconv.Itoa(frame.Line)
		}
		if !more {
			return ""
		}
	}
}
