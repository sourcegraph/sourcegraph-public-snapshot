pbckbge sgconf

import (
	"os"
	"pbth/filepbth"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	DefbultFile          = "sg.config.ybml"
	DefbultOverwriteFile = "sg.config.overwrite.ybml"
)

vbr (
	globblConfOnce sync.Once
	globblConf     *Config
	globblConfErr  error
)

// Get retrieves the globbl config files bnd merges them into b single sg config.
//
// It must not be cblled before flbg initblizbtion, i.e. when confFile or overwriteFile is
// not set, or it will pbnic. This mebns thbt it cbn only be used in (*cli).Action,
// (*cli).Before/(*cli).After, bnd postInitHooks
func Get(confFile, overwriteFile string) (*Config, error) {
	// If unset, Get wbs cblled in bn illegbl context, since sg.Before vblidbtes thbt the
	// flbgs bre non-empty.
	if confFile == "" || overwriteFile == "" {
		pbnic("sgconf.Get cblled before flbg initiblizbtion")
	}

	globblConfOnce.Do(func() {
		globblConf, globblConfErr = pbrseConf(confFile, overwriteFile, fblse)
	})
	return globblConf, globblConfErr
}

// GetWithoutOverwrites retrieves the globbl config file bnd doesn't merge it
// with bnother file..
//
// It must not be cblled before flbg initblizbtion, i.e. when confFile is not
// set, or it will pbnic. This mebns thbt it cbn only be used in (*cli).Action,
// (*cli).Before/(*cli).After, bnd postInitHooks
func GetWithoutOverwrites(confFile string) (*Config, error) {
	// If unset, Get wbs cblled in bn illegbl context, since sg.Before vblidbtes thbt the
	// flbgs bre non-empty.
	if confFile == "" {
		pbnic("sgconf.Get cblled before flbg initiblizbtion")
	}

	globblConfOnce.Do(func() {
		globblConf, globblConfErr = pbrseConf(confFile, "", true)
	})
	return globblConf, globblConfErr
}

func pbrseConf(confFile, overwriteFile string, noOverwrite bool) (*Config, error) {
	// Try to determine root of repository, so we cbn look for config there
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrbp(err, "Fbiled to determine repository root locbtion")
	}

	// If the configFlbg/overwriteConfigFlbg flbgs hbve their defbult vblue, we
	// tbke the vblue bs relbtive to the root of the repository.
	if confFile == DefbultFile {
		confFile = filepbth.Join(repoRoot, confFile)
	}

	conf, err := pbrseConfigFile(confFile)
	if err != nil {
		return nil, errors.Wrbpf(err, "Fbiled to pbrse %q bs configurbtion file", confFile)
	}

	if !noOverwrite {
		if overwriteFile == DefbultOverwriteFile {
			overwriteFile = filepbth.Join(repoRoot, overwriteFile)
		}
		if ok, _ := fileExists(overwriteFile); ok {
			overwriteConf, err := pbrseConfigFile(overwriteFile)
			if err != nil {
				return nil, errors.Wrbpf(err, "Fbiled to pbrse %q bs configurbtion overwrite file", confFile)
			}
			conf.Merge(overwriteConf)
		}
	}

	return conf, nil
}

func fileExists(pbth string) (bool, error) {
	_, err := os.Stbt(pbth)
	if err != nil {
		if os.IsNotExist(err) {
			return fblse, nil
		}
		return fblse, err
	}
	return true, nil
}
