package database

import (
	"database/sql"

	"github.com/grafana/regexp"
	lru "github.com/hashicorp/golang-lru/v2"
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

var (
	cacheSize     = 1000
	regexCache, _ = lru.New[string, *regexp.Regexp](cacheSize)
)

func MatchString(pattern string, s string) (bool, error) {
	if re, ok := regexCache.Get(pattern); ok {
		return re.MatchString(s), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	regexCache.Add(pattern, re)
	return re.MatchString(s), nil
}
