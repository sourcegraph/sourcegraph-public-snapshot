package hook

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/internal/configurable"
	"github.com/sourcegraph/log/internal/sinkcores/outputcore"
	"github.com/sourcegraph/log/output"
)

type writerSyncerAdapter struct{ io.Writer }

func (writerSyncerAdapter) Sync() error { return nil }

// Writer hooks receiver to rendered log output at level in the requested format,
// typically one of 'json' or 'console'.
func Writer(logger log.Logger, receiver io.Writer, level log.Level, format output.Format) log.Logger {
	cl := configurable.Cast(logger)

	// Adapt to WriteSyncer in case receiver doesn't implement it
	var writeSyncer zapcore.WriteSyncer
	if ws, ok := receiver.(zapcore.WriteSyncer); ok {
		writeSyncer = ws
	} else {
		writeSyncer = writerSyncerAdapter{receiver}
	}

	core := outputcore.NewCore(writeSyncer, level.Parse(), format, zap.SamplingConfig{}, nil, false)
	return cl.WithCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, core)
	})
}
