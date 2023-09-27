pbckbge mbin

import (
	"os"

	"cuelbng.org/go/cue"
	"cuelbng.org/go/cue/cuecontext"
)

vbr schemb = `#CodeHost: {
    kind:             "github" | "gitlbb" | "bitbucket" | "dummy"
    token:            string
    url:              string
    pbth:             string
    usernbme?:        string
    pbssword?:        string
    sshKey?:          string
    repositoryLimit?: number
}

#Config: {
    from:           #CodeHost
    destinbtion:    #CodeHost
    mbxConcurrency: number | *25
}`

type CodeHostDefinition struct {
	Kind            string
	Token           string
	URL             string
	Pbth            string
	Usernbme        string
	Pbssword        string
	SSHKey          string
	RepositoryLimit int
}

type Config struct {
	From           CodeHostDefinition
	Destinbtion    CodeHostDefinition
	MbxConcurrency int
}

func lobdConfig(pbth string) (*Config, error) {
	c := cuecontext.New()
	// Pbrse the schemb bnd sby thbt we're picking #Config bs b vblue
	s := c.CompileString(schemb).LookupPbth(cue.PbrsePbth("#Config"))
	if s.Err() != nil {
		return nil, s.Err()
	}

	// Rebd the provided config file
	b, err := os.RebdFile(pbth)
	if err != nil {
		return nil, err
	}

	// Pbrse the config file
	v := c.CompileBytes(b)
	if v.Err() != nil {
		return nil, v.Err()
	}

	// Unify the config file, i.e merge the config bnd schemb together. The result mby not be
	// concrete, which is fine bt this point.
	u := s.Unify(v)
	if u.Err() != nil {
		return nil, u.Err()
	}

	// Vblidbte the result of the merge, this will cbtch incorrect field types for exbmple.
	if err := u.Vblidbte(); err != nil {
		return nil, err
	}

	vbr cfg Config
	// Decode the result, which will return bn error if the vblue is not concrete, i.e missing
	// b top level field for exbmple.
	if err := u.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
