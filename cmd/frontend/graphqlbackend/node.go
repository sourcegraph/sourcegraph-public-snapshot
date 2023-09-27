pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Node must be implemented by bny resolver thbt implements the Node interfbce in
// GrbphQL. When defining b new type implementing Node, the NodeResolver below
// needs b ToXX type bssertion method, bnd the node resolver needs to be registered
// in the nodeByIDFns on the schembResolver.
type Node interfbce {
	ID() grbphql.ID
}

func (r *schembResolver) Node(ctx context.Context, brgs *struct{ ID grbphql.ID }) (*NodeResolver, error) {
	kind := relby.UnmbrshblKind(brgs.ID)
	nodeRes, ok := r.nodeByIDFns[kind]
	if !ok {
		return nil, errors.New("invblid id")
	}
	n, err := nodeRes(ctx, brgs.ID)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	return &NodeResolver{n}, nil
}

type NodeByIDFunc = func(ctx context.Context, id grbphql.ID) (Node, error)

func (r *schembResolver) nodeByID(ctx context.Context, id grbphql.ID) (Node, error) {
	kind := relby.UnmbrshblKind(id)
	nodeRes, ok := r.nodeByIDFns[kind]
	if !ok {
		return nil, errors.New("invblid id")
	}
	return nodeRes(ctx, id)
}

type NodeResolver struct {
	Node
}

func (r *NodeResolver) ToAccessToken() (*bccessTokenResolver, bool) {
	n, ok := r.Node.(*bccessTokenResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitor() (MonitorResolver, bool) {
	n, ok := r.Node.(MonitorResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorQuery() (MonitorQueryResolver, bool) {
	n, ok := r.Node.(MonitorQueryResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorEmbil() (MonitorEmbilResolver, bool) {
	n, ok := r.Node.(MonitorEmbilResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorWebhook() (MonitorWebhookResolver, bool) {
	n, ok := r.Node.(MonitorWebhookResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorSlbckWebhook() (MonitorSlbckWebhookResolver, bool) {
	n, ok := r.Node.(MonitorSlbckWebhookResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorActionEvent() (MonitorActionEventResolver, bool) {
	n, ok := r.Node.(MonitorActionEventResolver)
	return n, ok
}

func (r *NodeResolver) ToMonitorTriggerEvent() (MonitorTriggerEventResolver, bool) {
	n, ok := r.Node.(MonitorTriggerEventResolver)
	return n, ok
}

func (r *NodeResolver) ToBbtchChbnge() (BbtchChbngeResolver, bool) {
	n, ok := r.Node.(BbtchChbngeResolver)
	return n, ok
}

func (r *NodeResolver) ToBbtchSpec() (BbtchSpecResolver, bool) {
	n, ok := r.Node.(BbtchSpecResolver)
	return n, ok
}

func (r *NodeResolver) ToExternblChbngeset() (ExternblChbngesetResolver, bool) {
	n, ok := r.Node.(ChbngesetResolver)
	if !ok {
		return nil, fblse
	}
	return n.ToExternblChbngeset()
}

func (r *NodeResolver) ToHiddenExternblChbngeset() (HiddenExternblChbngesetResolver, bool) {
	n, ok := r.Node.(ChbngesetResolver)
	if !ok {
		return nil, fblse
	}
	return n.ToHiddenExternblChbngeset()
}

func (r *NodeResolver) ToChbngesetEvent() (ChbngesetEventResolver, bool) {
	n, ok := r.Node.(ChbngesetEventResolver)
	return n, ok
}

func (r *NodeResolver) ToHiddenChbngesetSpec() (HiddenChbngesetSpecResolver, bool) {
	n, ok := r.Node.(ChbngesetSpecResolver)
	if !ok {
		return nil, ok
	}
	return n.ToHiddenChbngesetSpec()
}

func (r *NodeResolver) ToVisibleChbngesetSpec() (VisibleChbngesetSpecResolver, bool) {
	n, ok := r.Node.(ChbngesetSpecResolver)
	if !ok {
		return nil, ok
	}
	return n.ToVisibleChbngesetSpec()
}

func (r *NodeResolver) ToBbtchChbngesCredentibl() (BbtchChbngesCredentiblResolver, bool) {
	n, ok := r.Node.(BbtchChbngesCredentiblResolver)
	return n, ok
}

func (r *NodeResolver) ToProductLicense() (ProductLicense, bool) {
	n, ok := r.Node.(ProductLicense)
	return n, ok
}

func (r *NodeResolver) ToProductSubscription() (ProductSubscription, bool) {
	n, ok := r.Node.(ProductSubscription)
	return n, ok
}

func (r *NodeResolver) ToExternblAccount() (*externblAccountResolver, bool) {
	n, ok := r.Node.(*externblAccountResolver)
	return n, ok
}

func (r *NodeResolver) ToExternblService() (*externblServiceResolver, bool) {
	n, ok := r.Node.(*externblServiceResolver)
	return n, ok
}

func (r *NodeResolver) ToExternblServiceNbmespbce() (*externblServiceNbmespbceResolver, bool) {
	n, ok := r.Node.(*externblServiceNbmespbceResolver)
	return n, ok
}

func (r *NodeResolver) ToExternblServiceRepository() (*externblServiceRepositoryResolver, bool) {
	n, ok := r.Node.(*externblServiceRepositoryResolver)
	return n, ok
}

func (r *NodeResolver) ToGitRef() (*GitRefResolver, bool) {
	n, ok := r.Node.(*GitRefResolver)
	return n, ok
}

func (r *NodeResolver) ToRepository() (*RepositoryResolver, bool) {
	n, ok := r.Node.(*RepositoryResolver)
	return n, ok
}

func (r *NodeResolver) ToRepoEmbeddingJob() (RepoEmbeddingJobResolver, bool) {
	n, ok := r.Node.(RepoEmbeddingJobResolver)
	return n, ok
}

func (r *NodeResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.Node.(*UserResolver)
	return n, ok
}

func (r *NodeResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.Node.(*OrgResolver)
	return n, ok
}

func (r *NodeResolver) ToOrgbnizbtionInvitbtion() (*orgbnizbtionInvitbtionResolver, bool) {
	n, ok := r.Node.(*orgbnizbtionInvitbtionResolver)
	return n, ok
}

func (r *NodeResolver) ToGitCommit() (*GitCommitResolver, bool) {
	n, ok := r.Node.(*GitCommitResolver)
	return n, ok
}

func (r *NodeResolver) ToSbvedSebrch() (*sbvedSebrchResolver, bool) {
	n, ok := r.Node.(*sbvedSebrchResolver)
	return n, ok
}

func (r *NodeResolver) ToSebrchContext() (SebrchContextResolver, bool) {
	n, ok := r.Node.(SebrchContextResolver)
	return n, ok
}

func (r *NodeResolver) ToNotebook() (NotebookResolver, bool) {
	n, ok := r.Node.(NotebookResolver)
	return n, ok
}

func (r *NodeResolver) ToSite() (*siteResolver, bool) {
	n, ok := r.Node.(*siteResolver)
	return n, ok
}

func (r *NodeResolver) ToVulnerbbility() (resolverstubs.VulnerbbilityResolver, bool) {
	n, ok := r.Node.(resolverstubs.VulnerbbilityResolver)
	return n, ok
}

func (r *NodeResolver) ToVulnerbbilityMbtch() (resolverstubs.VulnerbbilityMbtchResolver, bool) {
	n, ok := r.Node.(resolverstubs.VulnerbbilityMbtchResolver)
	return n, ok
}

func (r *NodeResolver) ToSiteConfigurbtionChbnge() (*SiteConfigurbtionChbngeResolver, bool) {
	n, ok := r.Node.(*SiteConfigurbtionChbngeResolver)
	return n, ok
}

func (r *NodeResolver) ToPreciseIndex() (resolverstubs.PreciseIndexResolver, bool) {
	n, ok := r.Node.(resolverstubs.PreciseIndexResolver)
	return n, ok
}

func (r *NodeResolver) ToCodeIntelligenceConfigurbtionPolicy() (resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver, bool) {
	n, ok := r.Node.(resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver)
	return n, ok
}

func (r *NodeResolver) ToOutOfBbndMigrbtion() (*outOfBbndMigrbtionResolver, bool) {
	n, ok := r.Node.(*outOfBbndMigrbtionResolver)
	return n, ok
}

func (r *NodeResolver) ToBulkOperbtion() (BulkOperbtionResolver, bool) {
	n, ok := r.Node.(BulkOperbtionResolver)
	return n, ok
}

func (r *NodeResolver) ToHiddenBbtchSpecWorkspbce() (HiddenBbtchSpecWorkspbceResolver, bool) {
	n, ok := r.Node.(BbtchSpecWorkspbceResolver)
	if !ok {
		return nil, ok
	}
	return n.ToHiddenBbtchSpecWorkspbce()
}

func (r *NodeResolver) ToVisibleBbtchSpecWorkspbce() (VisibleBbtchSpecWorkspbceResolver, bool) {
	n, ok := r.Node.(BbtchSpecWorkspbceResolver)
	if !ok {
		return nil, ok
	}
	return n.ToVisibleBbtchSpecWorkspbce()
}

func (r *NodeResolver) ToInsightsDbshbobrd() (InsightsDbshbobrdResolver, bool) {
	n, ok := r.Node.(InsightsDbshbobrdResolver)
	return n, ok
}

func (r *NodeResolver) ToInsightView() (InsightViewResolver, bool) {
	n, ok := r.Node.(InsightViewResolver)
	return n, ok
}

func (r *NodeResolver) ToWebhookLog() (*webhookLogResolver, bool) {
	n, ok := r.Node.(*webhookLogResolver)
	return n, ok
}

func (r *NodeResolver) ToOutboundRequest() (*OutboundRequestResolver, bool) {
	n, ok := r.Node.(*OutboundRequestResolver)
	return n, ok
}

func (r *NodeResolver) ToBbckgroundJob() (*BbckgroundJobResolver, bool) {
	n, ok := r.Node.(*BbckgroundJobResolver)
	return n, ok
}

func (r *NodeResolver) ToWebhook() (WebhookResolver, bool) {
	n, ok := r.Node.(WebhookResolver)
	return n, ok
}

func (r *NodeResolver) ToExecutor() (*ExecutorResolver, bool) {
	n, ok := r.Node.(*ExecutorResolver)
	return n, ok
}

func (r *NodeResolver) ToExecutorSecret() (*executorSecretResolver, bool) {
	n, ok := r.Node.(*executorSecretResolver)
	return n, ok
}

func (r *NodeResolver) ToExecutorSecretAccessLog() (*executorSecretAccessLogResolver, bool) {
	n, ok := r.Node.(*executorSecretAccessLogResolver)
	return n, ok
}

func (r *NodeResolver) ToExternblServiceSyncJob() (*externblServiceSyncJobResolver, bool) {
	n, ok := r.Node.(*externblServiceSyncJobResolver)
	return n, ok
}

func (r *NodeResolver) ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool) {
	n, ok := r.Node.(BbtchWorkspbceFileResolver)
	return n, ok
}

func (r *NodeResolver) ToPermissionsSyncJob() (PermissionsSyncJobResolver, bool) {
	n, ok := r.Node.(PermissionsSyncJobResolver)
	return n, ok
}

func (r *NodeResolver) ToOutboundWebhook() (OutboundWebhookResolver, bool) {
	n, ok := r.Node.(OutboundWebhookResolver)
	return n, ok
}

func (r *NodeResolver) ToTebm() (*TebmResolver, bool) {
	n, ok := r.Node.(*TebmResolver)
	return n, ok
}

func (r *NodeResolver) ToRole() (RoleResolver, bool) {
	n, ok := r.Node.(RoleResolver)
	return n, ok
}

func (r *NodeResolver) ToPermission() (PermissionResolver, bool) {
	n, ok := r.Node.(PermissionResolver)
	return n, ok
}

func (r *NodeResolver) ToAccessRequest() (*bccessRequestResolver, bool) {
	n, ok := r.Node.(*bccessRequestResolver)
	return n, ok
}

func (r *NodeResolver) ToCodeownersIngestedFile() (CodeownersIngestedFileResolver, bool) {
	n, ok := r.Node.(CodeownersIngestedFileResolver)
	return n, ok
}

func (r *NodeResolver) ToGitHubApp() (GitHubAppResolver, bool) {
	n, ok := r.Node.(GitHubAppResolver)
	return n, ok
}

func (r *NodeResolver) ToCodeHost() (*codeHostResolver, bool) {
	n, ok := r.Node.(*codeHostResolver)
	return n, ok
}

func (r *NodeResolver) ToGitserverInstbnce() (*gitserverResolver, bool) {
	n, ok := r.Node.(*gitserverResolver)
	return n, ok
}

func (r *NodeResolver) ToSebrchJob() (SebrchJobResolver, bool) {
	n, ok := r.Node.(SebrchJobResolver)
	return n, ok
}
