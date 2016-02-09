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
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
)

// Config specifies what authorization checks should be performed by
// wrapped services (obtained by calling Wrap).
type Config struct {
	AllowAnonymousReaders bool // whether to allow unauthenticated actors (non-logged-in users)

	DebugLog bool // debug log each call to authenticate
}

// Authenticate returns a non-nil error if authentication failed
// (according to c's config).
func (c *Config) Authenticate(ctx context.Context, accessLevel, label, repoURI string) (err error) {
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

	switch accessLevel {
	case "none":
		// Always allow methods that are called to establish
		// authentication; otherwise users wouldn't be able ever log in.
		return nil
	case "read":
		return accesscontrol.VerifyUserHasReadAccess(ctx, label, repoURI)
	case "write":
		return accesscontrol.VerifyUserHasWriteAccess(ctx, label, repoURI)
	case "admin":
		return accesscontrol.VerifyUserHasAdminAccess(ctx, label)
	}

	return grpc.Errorf(codes.Unauthenticated, "%s requires auth", label)
}
