// Package ctxdebug contains debug helpers for net/context and gRPC.
package ctxdebug

import (
	"log"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"src.sourcegraph.com/sourcegraph/auth/accesstoken"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// LogAuth logs the incoming and outgoing auth data stored in ctx.
func LogAuth(label string, ctx context.Context) {
	md, _ := metadata.FromContext(ctx)
	var incAuth string
	if md != nil && md["authorization"] != nil {
		incAuth = md["authorization"][0]
	}
	incomingTok, _ := accesstoken.UnsafeParseNoVerify(strings.TrimPrefix(incAuth, "Bearer "))
	if incomingTok != nil {
		incomingTok.Raw = ""
	}
	var outgoingTok *jwt.Token
	if cred := sourcegraph.CredentialsFromContext(ctx); cred != nil {
		md2, _ := (oauth.TokenSource{TokenSource: cred}).GetRequestMetadata(ctx)
		outgoingTok, _ = accesstoken.UnsafeParseNoVerify(strings.TrimPrefix(md2["authorization"], "Bearer "))
		if outgoingTok != nil {
			outgoingTok.Raw = ""
		}
	}
	log.Printf("#####%s##### Context auth\nIncoming: %+v\nOutgoing: %+v\n",
		label,
		incomingTok,
		outgoingTok,
	)
}
