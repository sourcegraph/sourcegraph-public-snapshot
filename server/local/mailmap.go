package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

func mapEmailsToUIDs(ctx context.Context, emailAddrs []string) (map[string]int, error) {
	m := make(map[string]int, len(emailAddrs))
	for _, email := range emailAddrs {
		userSpec, err := store.DirectoryFromContext(ctx).GetUserByEmail(ctx, email)
		if _, ok := err.(*store.UserNotFoundError); ok {
			continue
		} else if err != nil {
			return nil, err
		} else if userSpec == nil {
			continue
		}
		m[email] = int(userSpec.UID)
	}
	return m, nil
}
