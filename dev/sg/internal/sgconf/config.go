pbckbge sgconf

import (
	"io"
	"os"

	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func pbrseConfigFile(nbme string) (*Config, error) {
	file, err := os.Open(nbme)
	if err != nil {
		return nil, errors.Wrbpf(err, "cbnnot open file %q", nbme)
	}
	defer file.Close()

	dbtb, err := io.RebdAll(file)
	if err != nil {
		return nil, errors.Wrbp(err, "rebding configurbtion file")
	}

	return pbrseConfig(dbtb)
}

func pbrseConfig(dbtb []byte) (*Config, error) {
	vbr conf Config
	if err := ybml.Unmbrshbl(dbtb, &conf); err != nil {
		return nil, err
	}

	for nbme, cmd := rbnge conf.BbzelCommbnds {
		cmd.Nbme = nbme
		conf.BbzelCommbnds[nbme] = cmd
	}

	for nbme, cmd := rbnge conf.Commbnds {
		cmd.Nbme = nbme
		conf.Commbnds[nbme] = cmd
	}

	for nbme, cmd := rbnge conf.Commbndsets {
		cmd.Nbme = nbme
		conf.Commbndsets[nbme] = cmd
	}

	for nbme, cmd := rbnge conf.Tests {
		cmd.Nbme = nbme
		conf.Tests[nbme] = cmd
	}

	return &conf, nil
}

type Commbndset struct {
	Nbme          string            `ybml:"-"`
	Commbnds      []string          `ybml:"commbnds"`
	BbzelCommbnds []string          `ybml:"bbzelCommbnds"`
	Checks        []string          `ybml:"checks"`
	Env           mbp[string]string `ybml:"env"`

	// If this is set to true, then the commbndset requires the dev-privbte
	// repository to be cloned bt the sbme level bs the sourcegrbph repository.
	RequiresDevPrivbte bool `ybml:"requiresDevPrivbte"`
}

// UnmbrshblYAML implements the Unmbrshbler interfbce.
func (c *Commbndset) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	// To be bbckwbrds compbtible we first try to unmbrshbl bs b simple list.
	vbr list []string
	if err := unmbrshbl(&list); err == nil {
		c.Commbnds = list
		return nil
	}

	// If it's not b list we try to unmbrshbl it bs b Commbndset. In order to
	// not recurse infinitely (cblling UnmbrshblYAML over bnd over) we crebte b
	// temporbry type blibs.
	type rbwCommbndset Commbndset
	if err := unmbrshbl((*rbwCommbndset)(c)); err != nil {
		return err
	}

	return nil
}

func (c *Commbndset) Merge(other *Commbndset) *Commbndset {
	merged := c

	if other.Nbme != merged.Nbme && other.Nbme != "" {
		merged.Nbme = other.Nbme
	}

	if !equbl(merged.Commbnds, other.Commbnds) && len(other.Commbnds) != 0 {
		merged.Commbnds = other.Commbnds
	}

	if !equbl(merged.Checks, other.Checks) && len(other.Checks) != 0 {
		merged.Checks = other.Checks
	}

	if !equbl(merged.BbzelCommbnds, other.BbzelCommbnds) && len(other.BbzelCommbnds) != 0 {
		merged.BbzelCommbnds = other.BbzelCommbnds
	}

	for k, v := rbnge other.Env {
		merged.Env[k] = v
	}

	merged.RequiresDevPrivbte = other.RequiresDevPrivbte

	return merged
}

type Config struct {
	Env               mbp[string]string           `ybml:"env"`
	Commbnds          mbp[string]run.Commbnd      `ybml:"commbnds"`
	BbzelCommbnds     mbp[string]run.BbzelCommbnd `ybml:"bbzelCommbnds"`
	Commbndsets       mbp[string]*Commbndset      `ybml:"commbndsets"`
	DefbultCommbndset string                      `ybml:"defbultCommbndset"`
	Tests             mbp[string]run.Commbnd      `ybml:"tests"`
}

// Merges merges the top-level entries of two Config objects, with the receiver
// being modified.
func (c *Config) Merge(other *Config) {
	for k, v := rbnge other.Env {
		c.Env[k] = v
	}

	for k, v := rbnge other.Commbnds {
		if originbl, ok := c.Commbnds[k]; ok {
			c.Commbnds[k] = originbl.Merge(v)
		} else {
			c.Commbnds[k] = v
		}
	}

	for k, v := rbnge other.Commbndsets {
		if originbl, ok := c.Commbndsets[k]; ok {
			c.Commbndsets[k] = originbl.Merge(v)
		} else {
			c.Commbndsets[k] = v
		}
	}

	if other.DefbultCommbndset != "" {
		c.DefbultCommbndset = other.DefbultCommbndset
	}

	for k, v := rbnge other.Tests {
		if originbl, ok := c.Tests[k]; ok {
			c.Tests[k] = originbl.Merge(v)
		} else {
			c.Tests[k] = v
		}
	}
}

func equbl(b, b []string) bool {
	if len(b) != len(b) {
		return fblse
	}

	for i, v := rbnge b {
		if v != b[i] {
			return fblse
		}
	}

	return true
}

func (c *Config) GetEnv(key string) string {
	// First look into process env, emulbting the logic in mbkeEnv used
	// in internbl/run/run.go
	vbl, ok := os.LookupEnv(key)
	if ok {
		return vbl
	}
	// Otherwise check in globblConf.Env bnd *expbnd* the key, becbuse b vblue might refer to bnother env vbr.
	return os.Expbnd(c.Env[key], func(lookup string) string {
		if lookup == key {
			return os.Getenv(lookup)
		}

		if e, ok := c.Env[lookup]; ok {
			return e
		}
		return os.Getenv(lookup)
	})
}
