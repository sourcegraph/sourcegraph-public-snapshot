package local

import (
	"fmt"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

func (s *defs) ListClients(ctx context.Context, op *sourcegraph.DefsListClientsOp) (*sourcegraph.DefClientList, error) {
	def := op.Def

	refs, err := svc.Defs(ctx).ListRefs(ctx, &sourcegraph.DefsListRefsOp{
		Def: def,
		Opt: &sourcegraph.DefListRefsOptions{
			Authorship:  true,
			ListOptions: sourcegraph.ListOptions{PerPage: 25},
		},
	})
	if err != nil {
		return nil, err
	}

	clients := map[string]*sourcegraph.DefClient{}
	for _, ref := range refs.Refs {
		if ref.Authorship == nil {
			continue
		}
		if c, present := clients[ref.Authorship.AuthorEmail]; present {
			if c.LastCommitDate.Time().Before(ref.Authorship.LastCommitDate.Time()) {
				c.LastCommitDate = ref.Authorship.LastCommitDate
				c.LastCommitID = ref.Authorship.LastCommitID
			}
		} else {
			clients[ref.Authorship.AuthorEmail] = &sourcegraph.DefClient{
				Email: ref.Authorship.AuthorEmail,
				AuthorshipInfo: sourcegraph.AuthorshipInfo{
					AuthorEmail:    ref.Authorship.AuthorEmail,
					LastCommitDate: ref.Authorship.LastCommitDate,
					LastCommitID:   ref.Authorship.LastCommitID,
				},
			}
		}
	}

	// Map to UIDs.
	emailAddrs := make([]string, len(clients))
	i := 0
	for email := range clients {
		emailAddrs[i] = email
		i++
	}
	emailToUID, err := mapEmailsToUIDs(ctx, emailAddrs)
	if err != nil {
		return nil, err
	}
	var clients2 []*sourcegraph.DefClient // keyed on either UID (if mapped) or email
	for uid, emails := range mapUIDsToEmails(emailToUID) {
		var uidDC *sourcegraph.DefClient
		for _, email := range emails {
			if clients[email] == nil {
				panic(fmt.Sprintf("clients map has no entry for mapped email %q", email))
			}
			if uidDC == nil {
				uidDC = clients[email]
			} else {
				if uidDC.LastCommitDate.Time().Before(clients[email].LastCommitDate.Time()) {
					uidDC.LastCommitDate = clients[email].LastCommitDate
					uidDC.LastCommitID = clients[email].LastCommitID
				}
			}
			delete(clients, email)
		}
		uidDC.Email = ""
		uidDC.UID = int32(uid)
		clients2 = append(clients2, uidDC)
	}
	// Add DefClients for unmapped emails (all mapped DefClients have
	// been deleted from clients map).
	for _, dc := range clients {
		clients2 = append(clients2, dc)
	}

	// Remove AuthorEmail (for privacy; it's not obscured)
	for _, dc := range clients2 {
		dc.AuthorEmail = ""
	}

	return &sourcegraph.DefClientList{DefClients: clients2}, nil
}
