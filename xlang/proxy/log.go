package proxy

import (
	"encoding/json"
	"log"

	"github.com/fatih/color"
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
		"repo":    id.rootPath.Repo(),
		"rev":     id.rootPath.Rev(),
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
	switch lvl {
	case err:
		msg = color.New(color.BgRed).Sprint("ERROR") + " " + msg
		break
	case warn:
		msg = color.New(color.BgYellow).Sprint("WARN") + "  " + msg
		break
	case info:
		msg = color.New(color.BgCyan).Sprint("INFO") + "  " + msg
		break
	case debug:
		msg = "DEBUG " + msg
		break
	default:
		msg = color.New(color.Faint).Sprint("VERB" + string(lvl) + " " + msg)
		break
	}
	msg += " " + color.New(color.Faint).Sprint(metaToStr(id, extraMeta...))
	log.Println(msg)
}

// logError logs a message with LSP metadata with loglevel ERROR
func logError(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(err, msg, id, extraMeta...)
}

// logInfo logs a message with LSP metadata with loglevel INFO
func logInfo(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(info, msg, id, extraMeta...)
}

// logWarn logs a message with LSP metadata with loglevel WARN
func logWarn(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(warn, msg, id, extraMeta...)
}

// logDebug logs a message with LSP metadata with loglevel DEBUG
func logDebug(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel(debug, msg, id, extraMeta...)
}
