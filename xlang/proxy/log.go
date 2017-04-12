package proxy

import (
	"encoding/json"
	"log"
)

// logWithLevel logs a message with LSP metadata with a passed loglevel
func logWithLevel(lvl string, msg string, id contextID, extraMeta ...interface{}) {
	meta := map[string]interface{}{
		"mode":    id.mode,
		"session": id.session,
		"repo":    id.rootPath.Repo(),
		"rev":     id.rootPath.Rev(),
	}
	if len(extraMeta)%2 != 0 {
		log.Printf("%s %s [invalid meta]", lvl, msg)
		return
	}
	for i := 0; i < len(extraMeta); i += 2 {
		key, ok := extraMeta[i].(string)
		if !ok {
			log.Printf("%s %s [invalid meta]", lvl, msg)
			return
		}
		meta[key] = extraMeta[i+1]
	}
	metaBuf, err := json.Marshal(meta)
	if err != nil {
		log.Printf("%s %s [error parsing meta: %s]", lvl, msg, err)
		return
	}
	log.Printf("%s %s %s", lvl, msg, string(metaBuf))
}

// logError logs a message with LSP metadata with loglevel ERROR
func logError(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel("ERROR", msg, id, extraMeta...)
}

// logInfo logs a message with LSP metadata with loglevel INFO
func logInfo(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel("INFO ", msg, id, extraMeta...)
}

// logWarn logs a message with LSP metadata with loglevel WARN
func logWarn(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel("WARN ", msg, id, extraMeta...)
}

// logDebug logs a message with LSP metadata with loglevel DEBUG
func logDebug(msg string, id contextID, extraMeta ...interface{}) {
	logWithLevel("DEBUG", msg, id, extraMeta...)
}
