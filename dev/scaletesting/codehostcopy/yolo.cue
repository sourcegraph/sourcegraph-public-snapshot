{
	from: {
		kind:            "github"
		token:           "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"
		url:             "https://ghe-scaletesting.sgdev.org"
		path:            "main-org" // Use @myusername if you're targeting a user instead of an organization
		username:        "sourcegraph"
		password:        "W2Lht8EeLTJUwpuRT2EA"
		sshKey:          "/Users/tech/.ssh/id_rsa.pub"
		repositoryLimit: 1000
	}
	destination: {
		kind:     "dummy"
		token:    "1235"
		url:      "https://$GITLAB_HOST"
		path:     "baz"
		username: "my-user"
		password: "my-password"
	}
}
