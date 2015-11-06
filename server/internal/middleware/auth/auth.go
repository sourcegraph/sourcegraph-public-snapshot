// Package auth provides functions for svc.Services that check
// authorization, if so configured.
package auth

import (
	"fmt"
	"log"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
)

// Config specifies what authorization checks should be performed by
// wrapped services (obtained by calling Wrap).
type Config struct {
	AllowAnonymousReaders bool // whether to allow unauthenticated actors (non-logged-in users)

	DebugLog bool // debug log each call to authenticate
}

// Authenticate returns a non-nil error if authentication failed
// (according to c's config).
func (c *Config) Authenticate(ctx context.Context, label string) (err error) {
	if c.DebugLog {
		defer func() {
			var outcome string
			if err == nil {
				outcome = "OK"
			} else {
				outcome = fmt.Sprintf("DENIED (%s)", err)

				md, _ := metadata.FromContext(ctx)
				log.Println("MD = ", md)

			}

			var info []string
			a := auth.ActorFromContext(ctx)
			if a.UID != 0 {
				info = append(info, fmt.Sprintf("uid %d", a.UID))
			} else {
				info = append(info, "no user")
			}
			if a.ClientID != "" {
				info = append(info, fmt.Sprintf("client %s", a.ClientID))
			} else {
				info = append(info, "no client")
			}
			if len(a.Scope) > 0 {
				info = append(info, fmt.Sprintf("%d scopes %v", len(a.Scope), a.Scope))
			}
			var infoStr string
			if len(info) > 0 {
				infoStr = "[" + strings.Join(info, "; ") + "]"
			}

			log.Printf("authn: %s: %s %s", label, outcome, infoStr)
		}()
	}

	// Always allow methods that are called to establish
	// authentication; otherwise users wouldn't be able ever log in.
	switch label {
	case "Auth.GetAccessToken", "Accounts.Create", "Auth.Identify", "Meta.Config", "Accounts.RequestPasswordReset", "Accounts.ResetPassword", "RegisteredClients.GetCurrent":
		return nil
	case "GraphUplink.Push", "GraphUplink.PushEvents":
		// This is for backwards compatibility with client instances that are running older versions
		// of sourcegraph (< v0.7.22).
		// TODO: remove this hack once clients upgrade to binaries having the new grpc-go API.
		return nil
	}

	if c.AllowAnonymousReaders || !authutil.ActiveFlags.HasUserAccounts() {
		return nil
	}
	if auth.IsAuthenticated(ctx) || len(auth.ActorFromContext(ctx).Scope) > 0 {
		return nil
	}

	return grpc.Errorf(codes.Unauthenticated, "%s requires auth", label)
}
