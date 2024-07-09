package types

// SourceAuthenticationStrategy defines the possible types of authentication strategy that can
// be used to authenticate a ChangesetSource for a changeset.
type SourceAuthenticationStrategy string

const (
	// SourceAuthenticationStrategyUserCredential is used to authenticate using a traditional PAT configured by
	//the user or site admin. This should be used for all code host interactions unless another authentication
	// strategy is explicitly required.
	SourceAuthenticationStrategyUserCredential SourceAuthenticationStrategy = "USER_CREDENTIAL"

	// SourceAuthenticationStrategyGitHubApp is used to authenticate using a GitHub App.
	SourceAuthenticationStrategyGitHubApp SourceAuthenticationStrategy = "GITHUB_APP"
)
