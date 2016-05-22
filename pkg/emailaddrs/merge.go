package emailaddrs

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Merge merges 2 lists of email addresses. It returns the final list
// of email addresses to use, given a starting list (old) and a
// just-fetched list (new). It properly merges them, respecting
// blacklisted addresses, updating guessed flags, etc.
func Merge(old, new []*sourcegraph.EmailAddr) []*sourcegraph.EmailAddr {
	mergeEmail := func(old, new *sourcegraph.EmailAddr) *sourcegraph.EmailAddr {
		if old == nil {
			return new
		}

		if new == nil {
			if old.Guessed || old.Blacklisted {
				// Guessed emails aren't expected to be found each time we
				// do a fetch, so don't consider their absence evidence
				// that they should not be associated as guessed emails
				// with the user anymore; persist the existing (old)
				// guess. Same for blacklisted emails.
				return old
			}

			// Otherwise, assume this email is no longer associated with the user.
			return nil
		}

		// OK, both old and new are non-nil. Now we just need to merge the field values.

		if old.Blacklisted && new.Guessed {
			// Don't de-blacklist, since a guess is not strong enough evidence that it should be de-blacklisted.
			return old
		}

		return new
	}

	type oldAndNew struct{ old, new *sourcegraph.EmailAddr }
	emailsOldAndNew := make(map[string]*oldAndNew, len(new))

	// add emails to map, keyed on case-insensitive email
	for _, a := range old {
		e := strings.ToLower(a.Email)
		if _, present := emailsOldAndNew[e]; !present {
			emailsOldAndNew[e] = &oldAndNew{}
		}
		emailsOldAndNew[e].old = a
	}
	for _, a := range new {
		e := strings.ToLower(a.Email)
		if _, present := emailsOldAndNew[e]; !present {
			emailsOldAndNew[e] = &oldAndNew{}
		}
		emailsOldAndNew[e].new = a
	}

	merged := []*sourcegraph.EmailAddr{}
	for _, a := range emailsOldAndNew {
		m := mergeEmail(a.old, a.new)
		if m != nil {
			m.Email = strings.ToLower(m.Email)
			merged = append(merged, m)
		}
	}

	return merged
}
