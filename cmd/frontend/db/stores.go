package db

var (
	AccessTokens              = &accessTokens{}
	ExternalServices          = &ExternalServicesStore{}
	DiscussionThreads         = &discussionThreads{}
	DiscussionComments        = &discussionComments{}
	DiscussionMailReplyTokens = &discussionMailReplyTokens{}
	Repos                     = &repos{}
	Phabricator               = &phabricator{}
	QueryRunnerState          = &queryRunnerState{}
	Orgs                      = &orgs{}
	OrgMembers                = &orgMembers{}
	RecentSearches            = &recentSearches{}
	SavedSearches             = &savedSearches{}
	Settings                  = &settings{}
	Users                     = &users{}
	UserEmails                = &userEmails{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}
)
