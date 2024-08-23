package stderr

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Open() (zapcore.WriteSyncer, error) {
	errSink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}
	return errSink, nil
}
