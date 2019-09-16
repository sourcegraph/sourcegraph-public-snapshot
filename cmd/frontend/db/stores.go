package db

var (
	AccessTokens              = &accessTokens{}
	ExternalServices          = &ExternalServicesStore{}
	DefaultRepos              = &defaultRepos{}
	DiscussionThreads         = &discussionThreads{}
	DiscussionComments        = &discussionComments{}
	DiscussionMailReplyTokens = &discussionMailReplyTokens{}
	Repos                     = &repos{}
	Phabricator               = &phabricator{}
	QueryRunnerState          = &queryRunnerState{}
	Orgs                      = &orgs{}
	OrgMembers                = &orgMembers{}
	SavedSearches             = &savedSearches{}
	Settings                  = &settings{}
	Users                     = &users{}
	UserEmails                = &userEmails{}
	EventLogs                 = &eventLogs{}

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}
)
