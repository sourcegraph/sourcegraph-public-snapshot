package db

import "context"

type MockDiscussionMailReplyTokens struct {
	Generate func(ctx context.Context, userID int32, threadID int64) (string, error)
	Get      func(ctx context.Context, token string) (userID int32, threadID int64, err error)
}
