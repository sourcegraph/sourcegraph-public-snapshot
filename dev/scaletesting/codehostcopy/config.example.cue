{
	from: {
		kind:     "github"
		token:    "1235"
		url:      "https://$GHE_HOST/api/v3"
		path:     "baz"
		username: "my-user"
		password: "my-password"
	}
	destination: {
		kind:     "gitlab"
		token:    "1235"
		url:      "https://$GITLAB_HOST/api/v4"
		path:     "baz"
		username: "my-user"
		password: "my-password"
	}
}
