package db

var (
	AccessTokens              = &accessTokens{}
	ExternalServices          = &ExternalServicesStore{}
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

	SurveyResponses = &surveyResponses{}

	ExternalAccounts = &userExternalAccounts{}

	OrgInvitations = &orgInvitations{}
)
