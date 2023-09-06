package wrexec

import "context"

type (
	reason    string
	reasonKey int
)

func (r reason) ToString() string {
	return string(r)
}

const (
	commandReasonKey reasonKey = iota

	BatchChangeReason reason = "Batch Changes"
	JanitorReason     reason = "Janitor"
)

func SetCommandReason(ctx context.Context, r reason) context.Context {
	return context.WithValue(ctx, commandReasonKey, r)
}

func getCommandReason(ctx context.Context) reason {
	r := ctx.Value(commandReasonKey).(reason)
	return r
}
