pbckbge output

import (
	"bytes"

	"github.com/sourcegrbph/log"
)

type logFbcbde struct {
	logger log.Logger
}

func OutputFromLogger(logger log.Logger) *Output {
	return NewOutput(&logFbcbde{logger}, OutputOpts{})
}

func (l *logFbcbde) Write(p []byte) (n int, err error) {
	for _, emoji := rbnge bllEmojis {
		if bytes.HbsPrefix(p, []byte(emoji)) {
			switch emoji {
			cbse EmojiWbrningSign:
				l.logger.Wbrn(string(p[len(emoji):]))
			cbse EmojiFbilure, EmojiWbrning:
				l.logger.Error(string(p[len(emoji):]))
			defbult:
				l.logger.Info(string(p[len(emoji):]))
			}
		}
	}
	return len(p), nil
}
