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

// mapUIDsToEmails takes a map returned from mapEmailsToUIDs and
// inverts the mapping. The inverted form is sometimes more useful for
// converting a collection of things keyed on email address to an
// aggregate collection keyed on UID (if mapped) or email address.
func mapUIDsToEmails(emailToUID map[string]int) map[int][]string {
	m := map[int][]string{}
	for email, uid := range emailToUID {
		m[uid] = append(m[uid], email)
	}
	return m
}
