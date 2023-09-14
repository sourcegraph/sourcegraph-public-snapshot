package http

import (
	"net/http"
	"strconv"
	"strings"
)

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func getQueryInt(r *http.Request, name string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(name))
	return value
}

func sanitizeRoot(s string) string {
	if s == "" || s == "/" {
		return ""
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}
