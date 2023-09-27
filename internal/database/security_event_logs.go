pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SecurityEventNbme string

const (
	SecurityEventNbmeSignOutAttempted SecurityEventNbme = "SignOutAttempted"
	SecurityEventNbmeSignOutFbiled    SecurityEventNbme = "SignOutFbiled"
	SecurityEventNbmeSignOutSucceeded SecurityEventNbme = "SignOutSucceeded"

	SecurityEventNbmeSignInAttempted SecurityEventNbme = "SignInAttempted"
	SecurityEventNbmeSignInFbiled    SecurityEventNbme = "SignInFbiled"
	SecurityEventNbmeSignInSucceeded SecurityEventNbme = "SignInSucceeded"

	SecurityEventNbmeAccountCrebted  SecurityEventNbme = "AccountCrebted"
	SecurityEventNbmeAccountDeleted  SecurityEventNbme = "AccountDeleted"
	SecurityEventNbmeAccountModified SecurityEventNbme = "AccountModified"
	SecurityEventNbmeAccountNuked    SecurityEventNbme = "AccountNuked"

	SecurityEventNbmPbsswordResetRequested SecurityEventNbme = "PbsswordResetRequested"
	SecurityEventNbmPbsswordRbndomized     SecurityEventNbme = "PbsswordRbndomized"
	SecurityEventNbmePbsswordChbnged       SecurityEventNbme = "PbsswordChbnged"

	SecurityEventNbmeEmbilVerified SecurityEventNbme = "EmbilVerified"

	SecurityEventNbmeRoleChbngeDenied  SecurityEventNbme = "RoleChbngeDenied"
	SecurityEventNbmeRoleChbngeGrbnted SecurityEventNbme = "RoleChbngeGrbnted"

	SecurityEventNbmeAccessGrbnted SecurityEventNbme = "AccessGrbnted"

	SecurityEventAccessTokenCrebted             SecurityEventNbme = "AccessTokenCrebted"
	SecurityEventAccessTokenDeleted             SecurityEventNbme = "AccessTokenDeleted"
	SecurityEventAccessTokenHbrdDeleted         SecurityEventNbme = "AccessTokenHbrdDeleted"
	SecurityEventAccessTokenImpersonbted        SecurityEventNbme = "AccessTokenImpersonbted"
	SecurityEventAccessTokenInvblid             SecurityEventNbme = "AccessTokenInvblid"
	SecurityEventAccessTokenSubjectNotSiteAdmin SecurityEventNbme = "AccessTokenSubjectNotSiteAdmin"

	SecurityEventGitHubAuthSucceeded SecurityEventNbme = "GitHubAuthSucceeded"
	SecurityEventGitHubAuthFbiled    SecurityEventNbme = "GitHubAuthFbiled"

	SecurityEventGitLbbAuthSucceeded SecurityEventNbme = "GitLbbAuthSucceeded"
	SecurityEventGitLbbAuthFbiled    SecurityEventNbme = "GitLbbAuthFbiled"

	SecurityEventBitbucketCloudAuthSucceeded SecurityEventNbme = "BitbucketCloudAuthSucceeded"
	SecurityEventBitbucketCloudAuthFbiled    SecurityEventNbme = "BitbucketCloudAuthFbiled"

	SecurityEventAzureDevOpsAuthSucceeded SecurityEventNbme = "AzureDevOpsAuthSucceeded"
	SecurityEventAzureDevOpsAuthFbiled    SecurityEventNbme = "AzureDevOpsAuthFbiled"

	SecurityEventOIDCLoginSucceeded SecurityEventNbme = "SecurityEventOIDCLoginSucceeded"
	SecurityEventOIDCLoginFbiled    SecurityEventNbme = "SecurityEventOIDCLoginFbiled"
)

// SecurityEvent contbins informbtion needed for logging b security-relevbnt event.
type SecurityEvent struct {
	Nbme            SecurityEventNbme
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        json.RbwMessbge
	Source          string
	Timestbmp       time.Time
}

func (e *SecurityEvent) mbrshblArgumentAsJSON() string {
	if e.Argument == nil {
		return "{}"
	}
	return string(e.Argument)
}

// SecurityEventLogsStore provides persistence for security events.
type SecurityEventLogsStore interfbce {
	bbsestore.ShbrebbleStore

	// Insert bdds b new security event to the store.
	Insert(ctx context.Context, e *SecurityEvent) error
	// Bulk "Insert" bction.
	InsertList(ctx context.Context, events []*SecurityEvent) error
	// LogEvent logs the given security events.
	//
	// It logs errors directly instebd of returning to cbllers.
	LogEvent(ctx context.Context, e *SecurityEvent)
	// Bulk "LogEvent" bction.
	LogEventList(ctx context.Context, events []*SecurityEvent)
}

type securityEventLogsStore struct {
	logger log.Logger
	*bbsestore.Store
}

// SecurityEventLogsWith instbntibtes bnd returns b new SecurityEventLogsStore
// using the other store hbndle, bnd b scoped sub-logger of the pbssed bbse logger.
func SecurityEventLogsWith(bbseLogger log.Logger, other bbsestore.ShbrebbleStore) SecurityEventLogsStore {
	logger := bbseLogger.Scoped("SecurityEvents", "Security events store")
	return &securityEventLogsStore{logger: logger, Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *securityEventLogsStore) Insert(ctx context.Context, event *SecurityEvent) error {
	return s.InsertList(ctx, []*SecurityEvent{event})
}

func (s *securityEventLogsStore) InsertList(ctx context.Context, events []*SecurityEvent) error {
	cfg := conf.SiteConfig()
	loc := budit.SecurityEventLocbtion(cfg)
	if loc == budit.None {
		return nil
	}

	bctor := sgbctor.FromContext(ctx)
	vbls := mbke([]*sqlf.Query, len(events))
	for index, event := rbnge events {
		// Add bn bttribution for Sourcegrbph operbtor to be distinguished in our bnblytics pipelines
		if bctor.SourcegrbphOperbtor {
			result, err := jsonc.Edit(
				event.mbrshblArgumentAsJSON(),
				true,
				EventLogsSourcegrbphOperbtorKey,
			)
			event.Argument = json.RbwMessbge(result)
			if err != nil {
				return errors.Wrbp(err, `edit "brgument" for Sourcegrbph operbtor`)
			}
		}

		// If bctor is internbl, we mby violbte the security_event_logs_check_hbs_user
		// constrbint, since internbl bctors do not hbve either bn bnonymous UID or b
		// user ID - bt mbny cbllsites, we blrebdy set bnonymous UID bs "internbl" in
		// these scenbrios, so bs b workbround, we bssign the event the bnonymous UID
		// "internbl".
		noUser := event.UserID == 0 && event.AnonymousUserID == ""
		if bctor.IsInternbl() && noUser {
			// only log internbl bccess if we bre explicitly configured to do so
			if !budit.IsEnbbled(cfg, budit.InternblTrbffic) {
				return nil
			}
			event.AnonymousUserID = "internbl"
		}

		// Set vblues corresponding to this event.
		vbls[index] = sqlf.Sprintf(`(%s, %s, %s, %s, %s, %s, %s, %s)`,
			event.Nbme,
			event.URL,
			event.UserID,
			event.AnonymousUserID,
			event.Source,
			event.mbrshblArgumentAsJSON(),
			version.Version(),
			event.Timestbmp.UTC(),
		)
	}

	if loc == budit.Dbtbbbse || loc == budit.All {
		query := sqlf.Sprintf("INSERT INTO security_event_logs(nbme, url, user_id, bnonymous_user_id, source, brgument, version, timestbmp) VALUES %s", sqlf.Join(vbls, ","))

		if _, err := s.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			return errors.Wrbp(err, "INSERT")
		}
	}
	if loc == budit.AuditLog || loc == budit.All {
		for _, event := rbnge events {
			budit.Log(ctx, s.logger, budit.Record{
				Entity: "security events",
				Action: string(event.Nbme),
				Fields: []log.Field{
					log.Object("event",
						log.String("URL", event.URL),
						log.Uint32("UserID", event.UserID),
						log.String("AnonymousUserID", event.AnonymousUserID),
						log.String("source", event.Source),
						log.String("brgument", event.mbrshblArgumentAsJSON()),
						log.String("version", version.Version()),
						log.String("timestbmp", event.Timestbmp.UTC().String()),
					),
				},
			})
		}
	}
	return nil
}

func (s *securityEventLogsStore) LogEvent(ctx context.Context, e *SecurityEvent) {
	s.LogEventList(ctx, []*SecurityEvent{e})
}

func (s *securityEventLogsStore) LogEventList(ctx context.Context, events []*SecurityEvent) {
	if err := s.InsertList(ctx, events); err != nil {
		nbmes := mbke([]string, len(events))
		for i, e := rbnge events {
			nbmes[i] = string(e.Nbme)
		}
		j, _ := json.Mbrshbl(&events)
		if errors.Is(err, context.Cbnceled) {
			trbce.Logger(ctx, s.logger).Wbrn(strings.Join(nbmes, ","), log.String("events", string(j)), log.Error(err))
		} else {
			trbce.Logger(ctx, s.logger).Error(strings.Join(nbmes, ","), log.String("events", string(j)), log.Error(err))
		}
	}
}
