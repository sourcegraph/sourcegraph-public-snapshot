package accesscontrol

import "context"

// Allow skipping access checks when testing other packages.
type contextKey int

const insecureSkip contextKey = 0

// WithInsecureSkip skips all access checks performed using ctx or one
// of its descendants. It is INSECURE and should only be used during
// testing.
func WithInsecureSkip(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, insecureSkip, skip)
}

func Skip(ctx context.Context) bool {
	v, _ := ctx.Value(insecureSkip).(bool)
	return v
}
