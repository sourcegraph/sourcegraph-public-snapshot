package db

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/conf/confdb"
)

var (
	AccessTokens              = &accessTokens{}
	DiscussionThreads         = &discussionThreads{}
	DiscussionComments        = &discussionComments{}
	DiscussionMailReplyTokens = &discussionMailReplyTokens{}
	Repos                     = &repos{}
	Phabricator               = &phabricator{}
	SavedQueries              = &savedQueries{}
	Orgs                      = &orgs{}
	OrgMembers                = &orgMembers{}
	Settings                  = &settings{}
	Users                     = &users{}
	UserEmails                = &userEmails{}
	SiteConfig                = &siteConfig{}
	CertCache                 = &certCache{}
	SiteConfigurationFiles    = &confdb.SiteConfigurationFiles{Conn: func() *sql.DB { return dbconn.Global }}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}

	// GlobalDeps is a stub implementation of a global dependency index
	GlobalDeps GlobalDepsProvider = &globalDeps{}

	// Pkgs is a stub implementation of a global package metadata index
	Pkgs PkgsProvider = &pkgs{}
)
