package proxy

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

const (
	err   = 1
	warn  = 2
	info  = 3
	debug = 4
)

// metaToStr converts log metadata to a JSON string
func metaToStr(id contextID, extraMeta ...interface{}) string {
	meta := map[string]interface{}{
		"mode":    id.mode,
		"session": id.session,
		"repo":    id.rootURI.Repo(),
		"rev":     id.rootURI.Rev(),
	}
	if len(extraMeta)%2 != 0 {
		return "[invalid meta]"
	}
	for i := 0; i < len(extraMeta); i += 2 {
		key, ok := extraMeta[i].(string)
		if !ok {
			return "[invalid meta]"
		}
		meta[key] = extraMeta[i+1]
	}
	metaBuf, err := json.Marshal(meta)
	if err != nil {
		return "[error parsing meta: %s]"
	}
	return string(metaBuf)
}

// logWithLevel logs a message with LSP metadata with a passed loglevel
func logWithLevel(lvl int, msg string, id contextID, extraMeta ...interface{}) {
	var w io.Writer
	switch lvl {
	case err:
		msg = color.New(color.FgRed).Sprint("ERROR") + " " + msg
		w = env.ErrorOut
	case warn:
		msg = color.New(color.FgYellow).Sprint("WARN") + "  " + msg
		w = env.WarnOut
	case info:
		msg = color.New(color.FgCyan).Sprint("INFO") + "  " + msg
		w = env.InfoOut
	case debug:
		msg = color.New(color.Faint).Sprint("DEBUG") + " " + msg
		w = env.DebugOut
	default:
		msg = color.New(color.Faint).Sprint("VERB"+string(lvl)) + " " + msg
		w = env.DebugOut
	}
	msg += " " + color.New(color.Faint).Sprint(metaToStr(id, extraMeta...))
	fmt.Fprintln(w, msg)
}

// logError logs a message with LSP metadata with loglevel ERROR
func logError(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(err, msg, id, extraMeta...)
}

// logDebug logs a message with LSP metadata with loglevel DEBUG
func logDebug(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(debug, msg, id, extraMeta...)
}
