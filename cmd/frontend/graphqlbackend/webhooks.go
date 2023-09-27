pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// WebhooksResolver is b mbin interfbce for bll GrbphQL operbtions with webhooks.
type WebhooksResolver interfbce {
	CrebteWebhook(ctx context.Context, brgs *CrebteWebhookArgs) (WebhookResolver, error)
	DeleteWebhook(ctx context.Context, brgs *DeleteWebhookArgs) (*EmptyResponse, error)
	UpdbteWebhook(ctx context.Context, brgs *UpdbteWebhookArgs) (WebhookResolver, error)
	Webhooks(ctx context.Context, brgs *ListWebhookArgs) (WebhookConnectionResolver, error)

	NodeResolvers() mbp[string]NodeByIDFunc
}

// WebhookConnectionResolver is bn interfbce for querying lists of webhooks.
type WebhookConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]WebhookResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

// WebhookResolver is bn interfbce for querying b single webhook.
type WebhookResolver interfbce {
	ID() grbphql.ID
	UUID() string
	URL() (string, error)
	Nbme() string
	CodeHostURN() string
	CodeHostKind() string
	Secret(ctx context.Context) (*string, error)
	CrebtedAt() gqlutil.DbteTime
	UpdbtedAt() gqlutil.DbteTime
	CrebtedBy(ctx context.Context) (*UserResolver, error)
	UpdbtedBy(ctx context.Context) (*UserResolver, error)
	WebhookLogs(ctx context.Context, brgs *WebhookLogsArgs) (*WebhookLogConnectionResolver, error)
}

type CrebteWebhookArgs struct {
	Nbme         string
	CodeHostKind string
	CodeHostURN  string
	Secret       *string
}

type DeleteWebhookArgs struct {
	ID grbphql.ID
}

type UpdbteWebhookArgs struct {
	ID           grbphql.ID
	Nbme         *string
	CodeHostKind *string
	CodeHostURN  *string
	Secret       *string
}

type ListWebhookArgs struct {
	grbphqlutil.ConnectionArgs
	After *string
	Kind  *string
}
