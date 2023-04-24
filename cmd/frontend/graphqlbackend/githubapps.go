package graphqlbackend

import "context"

// This file just contains stub GraphQL resolvers and data types for Code Insights which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// enterprise/cmd/frontend/internal/auth/githubappauth/

type GitHubAppsResolver interface {
	// Queries

	// Mutations
	CreateGitHubApp(ctx context.Context, args *CreateGitHubAppArgs) (*int32, error)
}

type CreateGitHubAppInput struct {
	AppID        int32  `json:"appID"`
	BaseURL      string `json:"baseURL"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	PrivateKey   string `json:"privateKey"`
}

type CreateGitHubAppArgs struct {
	Input CreateGitHubAppInput `json:"input"`
}
