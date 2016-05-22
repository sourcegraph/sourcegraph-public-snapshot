package papertrail

import (
	"log"
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"

	"github.com/sourcegraph/go-papertrail/papertrail"
	"golang.org/x/net/context"
)

var (
	usePapertrail, _ = strconv.ParseBool(os.Getenv("SG_USE_PAPERTRAIL"))
	papertrailClient *papertrail.Client
)

func init() {
	if usePapertrail {
		token, err := papertrail.ReadToken()
		if err != nil {
			log.Printf("Warning: could not read Papertrail API token. Build logs will not be available. (Error was: %s.)", err)
			return
		}
		papertrailClient = papertrail.NewClient((&papertrail.TokenTransport{Token: token}).Client())
	}
}

func init() {
	if usePapertrail && papertrailClient != nil {
		serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
			ctx = newContext(ctx, papertrailClient)
			ctx = store.WithBuildLogs(ctx, &buildLogs{})
			return ctx, nil
		})
	}
}

type contextKey int

const (
	clientKey contextKey = iota
)

// newContext creates a new child context that holds a Papertrail API
// client.
func newContext(ctx context.Context, client *papertrail.Client) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

// client gets the context's Papertrail API client.
func client(ctx context.Context) *papertrail.Client {
	client, _ := ctx.Value(clientKey).(*papertrail.Client)
	if client == nil {
		panic("no Papertrail API client set in context")
	}
	return client
}
