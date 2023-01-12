package main

import (
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

var schema = `#CodeHost: {
    kind:             "github" | "gitlab" | "bitbucket" | "dummy"
    token:            string
    url:              string
    path:             string
    username?:        string
    password?:        string
    sshKey?:          string
    repositoryLimit?: number
}

#Config: {
    from:           #CodeHost
    destination:    #CodeHost
    maxConcurrency: number | *25
}`

type CodeHostDefinition struct {
	Kind            string
	Token           string
	URL             string
	Path            string
	Username        string
	Password        string
	SSHKey          string
	RepositoryLimit int
}

type Config struct {
	From           CodeHostDefinition
	Destination    CodeHostDefinition
	MaxConcurrency int
}

func loadConfig(path string) (*Config, error) {
	c := cuecontext.New()
	// Parse the schema and say that we're picking #Config as a value
	s := c.CompileString(schema).LookupPath(cue.ParsePath("#Config"))
	if s.Err() != nil {
		return nil, s.Err()
	}

	// Read the provided config file
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the config file
	v := c.CompileBytes(b)
	if v.Err() != nil {
		return nil, v.Err()
	}

	// Unify the config file, i.e merge the config and schema together. The result may not be
	// concrete, which is fine at this point.
	u := s.Unify(v)
	if u.Err() != nil {
		return nil, u.Err()
	}

	// Validate the result of the merge, this will catch incorrect field types for example.
	if err := u.Validate(); err != nil {
		return nil, err
	}

	var cfg Config
	// Decode the result, which will return an error if the value is not concrete, i.e missing
	// a top level field for example.
	if err := u.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
