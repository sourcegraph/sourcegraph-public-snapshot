{
	from: {
		kind:     "github"
		token:    "1235"
		url:      "https://$GHE_HOST"
		path:     "baz" // Use @myusername if you're targeting a user instead of an organization
		username: "my-user"
		password: "my-password"
	}
	destination: {
		kind:     "gitlab"
		token:    "1235"
		url:      "https://$GITLAB_HOST"
		path:     "baz"
		username: "my-user"
		password: "my-password"
	}
}
