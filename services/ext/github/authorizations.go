package github

import (
	"fmt"

	"context"
)

type Authorizations struct{}

func (s *Authorizations) Revoke(ctx context.Context, clientID, token string) error {
	resp, err := client(ctx).appAuthorizations.Revoke(clientID, token)
	if err != nil {
		return checkResponse(ctx, resp, err, fmt.Sprintf("github.Authorizations.Revoke %q <REDACTED>", clientID))
	}
	return nil
}
