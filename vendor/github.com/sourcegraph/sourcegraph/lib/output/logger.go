package output

import (
	"bytes"

	"github.com/sourcegraph/log"
)

type logFacade struct {
	logger log.Logger
}

func OutputFromLogger(logger log.Logger) *Output {
	return NewOutput(&logFacade{logger}, OutputOpts{})
}

func (l *logFacade) Write(p []byte) (n int, err error) {
	for _, emoji := range allEmojis {
		if bytes.HasPrefix(p, []byte(emoji)) {
			switch emoji {
			case EmojiWarningSign:
				l.logger.Warn(string(p[len(emoji):]))
			case EmojiFailure, EmojiWarning:
				l.logger.Error(string(p[len(emoji):]))
			default:
				l.logger.Info(string(p[len(emoji):]))
			}
		}
	}
	return len(p), nil
}
