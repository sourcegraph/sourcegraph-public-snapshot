package fs

import "golang.org/x/net/context"

type contextKey int

const (
	reposAbsPathKey contextKey = iota
)

// WithReposVFS creates a new child context that looks for
// repositories at the root of the given path.
func WithReposVFS(parent context.Context, absPath string) context.Context {
	return context.WithValue(parent, reposAbsPathKey, absPath)
}

// reposAbsPath returns the absolute path of the repository storage directory.
func reposAbsPath(ctx context.Context) string {
	return mustString(ctx, reposAbsPathKey)
}

func mustString(ctx context.Context, key contextKey) string {
	str, ok := ctx.Value(key).(string)
	if !ok {
		panic("no repos absolute path set in context")
	}
	return str
}
