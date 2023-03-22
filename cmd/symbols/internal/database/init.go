package database

import (
	"container/list"
	"database/sql"

	"github.com/grafana/regexp"
	"github.com/mattn/go-sqlite3"
)

func Init() {
	sql.Register("sqlite3_with_regexp",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("REGEXP", MatchString, true)
			},
		})
}

type entry struct {
	pattern string
	re      *regexp.Regexp
}

var (
	cacheSize  = 100 // Maximum number of entries in the cache
	regexCache = make(map[string]*list.Element)
	lruList    = list.New()
)

func MatchString(pattern string, s string) (matched bool, err error) {
	// Check if the compiled regular expression for the given pattern already exists in the cache
	if e, ok := regexCache[pattern]; ok {
		lruList.MoveToFront(e)
		return e.Value.(*entry).re.MatchString(s), nil
	}

	// If it doesn't exist, compile the regular expression and store it in the cache
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if len(regexCache) >= cacheSize {
		// If the cache has reached its maximum capacity, evict the least recently used entry
		oldest := lruList.Back()
		delete(regexCache, oldest.Value.(*entry).pattern)
		lruList.Remove(oldest)
	}

	// Add the new entry to the cache and move it to the front of the LRU list
	e := lruList.PushFront(&entry{pattern: pattern, re: re})
	regexCache[pattern] = e

	return re.MatchString(s), nil
}
