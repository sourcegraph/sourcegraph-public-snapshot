package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// recipientSpec identifies a recipient of a saved search notification. Exactly one of its fields is
// nonzero.
type recipientSpec struct {
	userID, orgID int32
}

func (r recipientSpec) String() string {
	if r.userID != 0 {
		return fmt.Sprintf("user %d", r.userID)
	}
	return fmt.Sprintf("org %d", r.orgID)
}

// recipient describes a recipient of a saved search notification and the type of notifications
// they're configured to receive.
type recipient struct {
	spec  recipientSpec // the recipient's identity
	email bool          // send an email to the recipient
	slack bool          // post a Slack message to the recipient
}

func (r *recipient) String() string {
	return fmt.Sprintf("{%s email:%v slack:%v}", r.spec, r.email, r.slack)
}

func (r recipient) subject() api.ConfigurationSubject {
	if r.spec.userID != 0 {
		return api.ConfigurationSubject{User: &r.spec.userID}
	}
	return api.ConfigurationSubject{Org: &r.spec.orgID}
}

// getNotificationRecipients retrieves the list of recipients who should receive notifications for
// events related to the saved search.
func getNotificationRecipients(ctx context.Context, spec api.SavedQueryIDSpec, query api.ConfigSavedQuery) ([]*recipient, error) {
	var recipients recipients

	// Notify the owner (user or org).
	switch {
	case spec.Subject.User != nil:
		recipients.add(recipient{
			spec:  recipientSpec{userID: *spec.Subject.User},
			email: query.Notify,
			slack: query.NotifySlack,
		})

	case spec.Subject.Org != nil:
		if query.Notify {
			// Email all org members.
			orgMembers, err := api.InternalClient.OrgsListUsers(ctx, *spec.Subject.Org)
			if err != nil {
				return nil, err
			}
			for _, userID := range orgMembers {
				recipients.add(recipient{
					spec:  recipientSpec{userID: userID},
					email: true,
				})
			}
		}

		recipients.add(recipient{
			spec:  recipientSpec{orgID: *spec.Subject.Org},
			slack: query.NotifySlack,
		})
	}

	return recipients, nil
}

type recipients []*recipient

// add adds the new recipient, merging it into an existing slice element if one already exists for
// the userID or orgID.
func (rs *recipients) add(r recipient) {
	for _, r2 := range *rs {
		if r.spec == r2.spec {
			// Merge into existing recipient.
			r2.email = r2.email || r.email
			r2.slack = r2.slack || r.slack
			return
		}
	}
	// Add new recipient.
	*rs = append(*rs, &r)
}

// get returns the recipient with the given spec, if any, or else nil.
func (rs recipients) get(s recipientSpec) *recipient {
	for _, r := range rs {
		if r.spec == s {
			return r
		}
	}
	return nil
}

// diffNotificationRecipients diffs old against new, returning the removed and added recipients. The
// same recipient identity may be returned in both the removed and added lists, if they changed the
// type of notifications they receive (e.g., unsubscribe from email, subscribe to Slack).
func diffNotificationRecipients(old, new recipients) (removed, added recipients) {
	diff := func(spec recipientSpec, old, new *recipient) (removed, added *recipient) {
		empty := recipient{spec: spec}
		if old == nil || *old == empty {
			return nil, new
		}
		if new == nil || *new == empty {
			return old, nil
		}
		if *old == *new {
			return nil, nil
		}
		removed = &recipient{
			spec:  spec,
			email: old.email && !new.email,
			slack: old.slack && !new.slack,
		}
		if *removed == empty {
			removed = nil
		}
		added = &recipient{
			spec:  spec,
			email: new.email && !old.email,
			slack: new.slack && !old.slack,
		}
		if *added == empty {
			added = nil
		}
		return removed, added
	}

	seen := map[recipientSpec]struct{}{}
	handle := func(spec recipientSpec, oldr, newr *recipient) {
		if _, seen := seen[spec]; seen {
			return
		}
		seen[spec] = struct{}{}
		removedr, addedr := diff(spec, oldr, newr)
		if removedr != nil {
			removed.add(*removedr)
		}
		if addedr != nil {
			added.add(*addedr)
		}
	}
	for _, oldr := range old {
		handle(oldr.spec, oldr, new.get(oldr.spec))
	}
	for _, newr := range new {
		handle(newr.spec, old.get(newr.spec), newr)
	}
	return removed, added
}
