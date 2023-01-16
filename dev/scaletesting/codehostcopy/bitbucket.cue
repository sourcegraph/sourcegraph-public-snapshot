{
	from: {
		kind:            "bitbucket"
		token:           "MzQwMzA4NDc1Njg5Oix4pY+pSW3S3PzswiFDbIa+2gl4"
		url:             "https://bitbucket.sgdev.org"
		path:            "batch-changes-testing" // Use @myusername if you're targeting a user instead of an organization
		username:        "milton"
		password:        "cM4N$Nr//HNyZhgysn88HC>eVm6n74"
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
