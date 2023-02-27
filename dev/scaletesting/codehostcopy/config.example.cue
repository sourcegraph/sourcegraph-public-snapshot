{
	from: {
		kind:            "github"
		token:           "1235"
		url:             "https://$GHE_HOST"
		path:            "baz" // Use @myusername if you're targeting a user instead of an organization
		username:        "my-user"
		password:        "my-password"
		sshKey:          "/path/to/.ssh/id_rsa.pub"
		repositoryLimit: 5000 // defines the amount of repositories to copy from the source. GHE fails to accurately list >50k private repos owned by the source, so define the exact amount here.
	}
	destination: {
		kind:     "gitlab"
		token:    "1235"
		url:      "https://$GITLAB_HOST"
		path:     "baz"
		username: "my-user"
		password: "my-password"
	}
	maxConcurrency: 25 // how many repositories can be cloned from the source and created at the destination concurrently
}
