package search

import (
	"log"
	"path"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/svc"
)

// Shorten transforms a list of tokens into shorter corresponding
// strings that resolve to the original list of longer tokens. It is
// used to display to the user what they actually could type to issue
// a specific query.
func Shorten(ctx context.Context, tokens []sourcegraph.Token) ([]string, error) {
	currentUserIsOrgMember := map[string]bool{}
	if login, ok := auth.LoginFromContext(ctx); ok {
		currentUserIsOrgMember[strings.ToLower(login)] = true

		orgs, err := svc.Orgs(ctx).List(ctx, &sourcegraph.OrgsListOp{Member: sourcegraph.UserSpec{Login: login}})
		if err != nil && grpc.Code(err) != codes.Unimplemented {
			// Pretend the user is in no orgs; it just means the resulting repo tokens will be longer.
			log.Printf("Shorten: ListOrgs(user=%q) error: %s", login, err)
		}
		if orgs != nil {
			for _, org := range orgs.Orgs {
				currentUserIsOrgMember[strings.ToLower(org.Login)] = true
			}
		}
	}

	shortened := make([]string, len(tokens))
	for i, tok := range tokens {
		switch tok := tok.(type) {
		case sourcegraph.RepoToken:
			tok.URI = strings.TrimPrefix(strings.TrimPrefix(tok.URI, "github.com/"), "sourcegraph.com/")
			if ownerEnd := strings.Index(tok.URI, "/"); ownerEnd != -1 && currentUserIsOrgMember[strings.ToLower(tok.URI[:ownerEnd])] {
				tok.URI = tok.URI[ownerEnd+1:]
			}
			shortened[i] = tok.Token()
		case sourcegraph.UnitToken:
			// This isn't guaranteed to work, but it usually will.
			tok.UnitType = ""
			tok.Name = path.Base(tok.Name)
			shortened[i] = tok.Name
		default:
			shortened[i] = tok.Token()
		}
	}
	return shortened, nil
}
