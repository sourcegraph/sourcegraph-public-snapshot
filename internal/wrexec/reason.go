package wrexec

import "context"

type (
	reason    string
	reasonKey int
)

const (
	commandReasonKey reasonKey = iota

	BatchChangeReason reason = "Batch Changes"
	JanitorReason     reason = "Janitor"
)

func SetCommandReason(ctx context.Context, r reason) context.Context {
	return context.WithValue(ctx, commandReasonKey, r)
}

func GetCommandReason(ctx context.Context) reason {
	r := ctx.Value(commandReasonKey).(reason)
	return r
}
