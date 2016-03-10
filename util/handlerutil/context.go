package handlerutil

type contextKey int

const (
	userKey contextKey = iota
	fullUserKey
	emailAddrKey
	repoFallbackKey
)
