package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	dbtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CodyGatewayDotcomUserResolver implements the GraphQL Query and Mutation fields related to Cody gateway users.
type CodyGatewayDotcomUserResolver struct {
	DB database.DB
}

func (r CodyGatewayDotcomUserResolver) DotcomCodyGatewayUserByToken(ctx context.Context, args *graphqlbackend.CodyGatewayUsersByAccessTokenArgs) (graphqlbackend.CodyGatewayUser, error) {
	// 🚨 SECURITY: Only site admins or the service accounts may check users.
	if err := serviceAccountOrSiteAdmin(ctx, r.DB, true); err != nil {
		return nil, err
	}
	dbTokens := newDBTokens(r.DB)
	userID, err := dbTokens.LookupDotcomUserIDByAccessToken(ctx, args.Token)
	if err != nil {
		return nil, err
	}

	user, err := r.DB.Users().GetByID(ctx, int32(userID))
	if err != nil {
		return nil, err
	}
	verified, err := r.DB.UserEmails().HasVerifiedEmail(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &dotcomCodyUserResolver{
		db:            r.DB,
		user:          user,
		verifiedEmail: verified,
	}, nil

}

type dotcomCodyUserResolver struct {
	db            database.DB
	user          *dbtypes.User
	verifiedEmail bool
}

func (u *dotcomCodyUserResolver) Username() string {
	return u.user.Username
}

func (u *dotcomCodyUserResolver) CodyGatewayAccess() graphqlbackend.CodyGatewayAccess {
	return &codyUserGatewayAccessResolver{
		db:            u.db,
		user:          u.user,
		verifiedEmail: u.verifiedEmail,
	}
}

type codyUserGatewayAccessResolver struct {
	db            database.DB
	user          *dbtypes.User
	verifiedEmail bool
}

func (r codyUserGatewayAccessResolver) Enabled() bool { return r.user.SiteAdmin || r.verifiedEmail }

func (r codyUserGatewayAccessResolver) ChatCompletionsRateLimit(ctx context.Context) (graphqlbackend.CodyGatewayRateLimit, error) {
	// If the user isn't enabled return a rate limit with 0 requests available
	if !r.Enabled() {
		return nil, nil
	}
	rateLimit, rateLimitSource, err := getRateLimit(ctx, r.db, r.user.ID, types.CompletionsFeatureChat)
	if err != nil {
		return nil, err
	}

	return &codyGatewayRateLimitResolver{
		feature:     types.CompletionsFeatureChat,
		actorID:     r.user.Username,
		actorSource: codygateway.ActorSourceDotcomUser,
		source:      rateLimitSource,
		v:           rateLimit,
	}, nil
}

func (r codyUserGatewayAccessResolver) CodeCompletionsRateLimit(ctx context.Context) (graphqlbackend.CodyGatewayRateLimit, error) {
	// If the user isn't enabled return a rate limit with 0 requests available
	if !r.Enabled() {
		return nil, nil
	}

	rateLimit, rateLimitSource, err := getRateLimit(ctx, r.db, r.user.ID, types.CompletionsFeatureCode)
	if err != nil {
		return nil, err
	}

	return &codyGatewayRateLimitResolver{
		feature:     types.CompletionsFeatureCode,
		actorID:     r.user.Username,
		actorSource: codygateway.ActorSourceDotcomUser,
		source:      rateLimitSource,
		v:           rateLimit,
	}, nil
}

func getRateLimit(ctx context.Context, db database.DB, userID int32, scope types.CompletionsFeature) (licensing.CodyGatewayRateLimit, graphqlbackend.CodyGatewayRateLimitSource, error) {
	var limit *int
	var err error
	source := graphqlbackend.CodyGatewayRateLimitSourceOverride

	switch scope {
	case types.CompletionsFeatureChat:
		limit, err = db.Users().GetChatCompletionsQuota(ctx, userID)
	case types.CompletionsFeatureCode:
		limit, err = db.Users().GetCodeCompletionsQuota(ctx, userID)
	default:
		return licensing.CodyGatewayRateLimit{}, graphqlbackend.CodyGatewayRateLimitSourcePlan, errors.Newf("unknown scope: %s", scope)
	}
	if err != nil {
		return licensing.CodyGatewayRateLimit{}, graphqlbackend.CodyGatewayRateLimitSourcePlan, err
	}
	if limit == nil {
		source = graphqlbackend.CodyGatewayRateLimitSourcePlan
		// Otherwise, fall back to the global limit.
		cfg := conf.Get()
		switch scope {
		case types.CompletionsFeatureChat:
			if cfg.Completions != nil && cfg.Completions.PerUserDailyLimit > 0 {
				limit = iPtr(cfg.Completions.PerUserDailyLimit)
			}
		case types.CompletionsFeatureCode:
			if cfg.Completions != nil && cfg.Completions.PerUserCodeCompletionsDailyLimit > 0 {
				limit = iPtr(cfg.Completions.PerUserCodeCompletionsDailyLimit)
			}
		default:
			return licensing.CodyGatewayRateLimit{}, graphqlbackend.CodyGatewayRateLimitSourcePlan, errors.Newf("unknown scope: %s", scope)
		}
	}
	if limit == nil {
		limit = iPtr(0)
	}
	return licensing.CodyGatewayRateLimit{
		AllowedModels:   allowedModels(scope),
		Limit:           int32(*limit),
		IntervalSeconds: 86400, // Daily limit
	}, source, nil
}

func allowedModels(scope types.CompletionsFeature) []string {
	switch scope {
	case types.CompletionsFeatureChat:
		return []string{"claude-v1"}
	case types.CompletionsFeatureCode:
		return []string{"claude-instant-v1"}
	default:
		return []string{}
	}
}

func iPtr(i int) *int {
	return &i
}
