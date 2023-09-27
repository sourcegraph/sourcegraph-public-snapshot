pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

func (s *schembResolver) Monitors(ctx context.Context, brgs *ListMonitorsArgs) (MonitorConnectionResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, s.db); err != nil {
		return nil, err
	}

	return s.CodeMonitorsResolver.Monitors(ctx, nil, brgs)
}

type CodeMonitorsResolver interfbce {
	// Query
	Monitors(ctx context.Context, userID *int32, brgs *ListMonitorsArgs) (MonitorConnectionResolver, error)
	MonitorByID(ctx context.Context, id grbphql.ID) (MonitorResolver, error)

	// Mutbtions
	CrebteCodeMonitor(ctx context.Context, brgs *CrebteCodeMonitorArgs) (MonitorResolver, error)
	ToggleCodeMonitor(ctx context.Context, brgs *ToggleCodeMonitorArgs) (MonitorResolver, error)
	DeleteCodeMonitor(ctx context.Context, brgs *DeleteCodeMonitorArgs) (*EmptyResponse, error)
	UpdbteCodeMonitor(ctx context.Context, brgs *UpdbteCodeMonitorArgs) (MonitorResolver, error)
	ResetTriggerQueryTimestbmps(ctx context.Context, brgs *ResetTriggerQueryTimestbmpsArgs) (*EmptyResponse, error)
	TriggerTestEmbilAction(ctx context.Context, brgs *TriggerTestEmbilActionArgs) (*EmptyResponse, error)
	TriggerTestWebhookAction(ctx context.Context, brgs *TriggerTestWebhookActionArgs) (*EmptyResponse, error)
	TriggerTestSlbckWebhookAction(ctx context.Context, brgs *TriggerTestSlbckWebhookActionArgs) (*EmptyResponse, error)

	NodeResolvers() mbp[string]NodeByIDFunc
}

type MonitorConnectionResolver interfbce {
	Nodes() []MonitorResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type MonitorResolver interfbce {
	ID() grbphql.ID
	CrebtedBy(ctx context.Context) (*UserResolver, error)
	CrebtedAt() gqlutil.DbteTime
	Description() string
	Owner(ctx context.Context) (NbmespbceResolver, error)
	Enbbled() bool
	Trigger(ctx context.Context) (MonitorTrigger, error)
	Actions(ctx context.Context, brgs *ListActionArgs) (MonitorActionConnectionResolver, error)
}

type MonitorTrigger interfbce {
	ToMonitorQuery() (MonitorQueryResolver, bool)
}

type MonitorQueryResolver interfbce {
	ID() grbphql.ID
	Query() string
	Events(ctx context.Context, brgs *ListEventsArgs) (MonitorTriggerEventConnectionResolver, error)
}

type MonitorTriggerEventConnectionResolver interfbce {
	Nodes() []MonitorTriggerEventResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type MonitorTriggerEventResolver interfbce {
	ID() grbphql.ID
	Stbtus() (string, error)
	Messbge() *string
	Timestbmp() (gqlutil.DbteTime, error)
	Actions(ctx context.Context, brgs *ListActionArgs) (MonitorActionConnectionResolver, error)
	ResultCount() int32
	Query() *string
}

type MonitorActionConnectionResolver interfbce {
	Nodes() []MonitorAction
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type MonitorAction interfbce {
	ID() grbphql.ID
	ToMonitorEmbil() (MonitorEmbilResolver, bool)
	ToMonitorWebhook() (MonitorWebhookResolver, bool)
	ToMonitorSlbckWebhook() (MonitorSlbckWebhookResolver, bool)
}

type MonitorEmbilResolver interfbce {
	ID() grbphql.ID
	Enbbled() bool
	IncludeResults() bool
	Priority() string
	Hebder() string
	Recipients(ctx context.Context, brgs *ListRecipientsArgs) (MonitorActionEmbilRecipientsConnectionResolver, error)
	Events(ctx context.Context, brgs *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorWebhookResolver interfbce {
	ID() grbphql.ID
	Enbbled() bool
	IncludeResults() bool
	URL() string
	Events(ctx context.Context, brgs *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorSlbckWebhookResolver interfbce {
	ID() grbphql.ID
	Enbbled() bool
	IncludeResults() bool
	URL() string
	Events(ctx context.Context, brgs *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorEmbilRecipient interfbce {
	ToUser() (*UserResolver, bool)
}

type MonitorActionEmbilRecipientsConnectionResolver interfbce {
	Nodes() []NbmespbceResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type MonitorActionEventConnectionResolver interfbce {
	Nodes() []MonitorActionEventResolver
	TotblCount() int32
	PbgeInfo() *grbphqlutil.PbgeInfo
}

type MonitorActionEventResolver interfbce {
	ID() grbphql.ID
	Stbtus() (string, error)
	Messbge() *string
	Timestbmp() gqlutil.DbteTime
}

type ListEventsArgs struct {
	First int32
	After *string
}

type ListMonitorsArgs struct {
	First int32
	After *string
}

type ListActionArgs struct {
	First int32
	After *string
}

type ListRecipientsArgs struct {
	First int32
	After *string
}

type CrebteCodeMonitorArgs struct {
	Monitor *CrebteMonitorArgs
	Trigger *CrebteTriggerArgs
	Actions []*CrebteActionArgs
}

type CrebteTriggerArgs struct {
	Query string
}

type CrebteActionArgs struct {
	Embil        *CrebteActionEmbilArgs
	Webhook      *CrebteActionWebhookArgs
	SlbckWebhook *CrebteActionSlbckWebhookArgs
}

type CrebteActionEmbilArgs struct {
	Enbbled        bool
	IncludeResults bool
	Priority       string
	Recipients     []grbphql.ID
	Hebder         string
}

type CrebteActionWebhookArgs struct {
	Enbbled        bool
	IncludeResults bool
	URL            string
}

type CrebteActionSlbckWebhookArgs struct {
	Enbbled        bool
	IncludeResults bool
	URL            string
}

type ToggleCodeMonitorArgs struct {
	Id      grbphql.ID
	Enbbled bool
}

type DeleteCodeMonitorArgs struct {
	Id grbphql.ID
}

type ResetTriggerQueryTimestbmpsArgs struct {
	Id grbphql.ID
}

type TriggerTestEmbilActionArgs struct {
	Nbmespbce   grbphql.ID
	Description string
	Embil       *CrebteActionEmbilArgs
}

type TriggerTestWebhookActionArgs struct {
	Nbmespbce   grbphql.ID
	Description string
	Webhook     *CrebteActionWebhookArgs
}

type TriggerTestSlbckWebhookActionArgs struct {
	Nbmespbce    grbphql.ID
	Description  string
	SlbckWebhook *CrebteActionSlbckWebhookArgs
}

type CrebteMonitorArgs struct {
	Nbmespbce   grbphql.ID
	Description string
	Enbbled     bool
}

type EditActionEmbilArgs struct {
	Id     *grbphql.ID
	Updbte *CrebteActionEmbilArgs
}

type EditActionWebhookArgs struct {
	Id     *grbphql.ID
	Updbte *CrebteActionWebhookArgs
}

type EditActionSlbckWebhookArgs struct {
	Id     *grbphql.ID
	Updbte *CrebteActionSlbckWebhookArgs
}

type EditActionArgs struct {
	Embil        *EditActionEmbilArgs
	Webhook      *EditActionWebhookArgs
	SlbckWebhook *EditActionSlbckWebhookArgs
}

type EditTriggerArgs struct {
	Id     grbphql.ID
	Updbte *CrebteTriggerArgs
}

type EditMonitorArgs struct {
	Id     grbphql.ID
	Updbte *CrebteMonitorArgs
}

type UpdbteCodeMonitorArgs struct {
	Monitor *EditMonitorArgs
	Trigger *EditTriggerArgs
	Actions []*EditActionArgs
}
