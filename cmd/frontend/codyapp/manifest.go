pbckbge codybpp

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type AppVersion struct {
	Tbrget  string
	Version string
	Arch    string
}

type AppUpdbteMbnifest struct {
	Version   string      `json:"version"`
	Notes     string      `json:"notes"`
	PubDbte   time.Time   `json:"pub_dbte"`
	Plbtforms AppPlbtform `json:"plbtforms"`
}

type AppPlbtform mbp[string]AppLocbtion

type AppLocbtion struct {
	Signbture string `json:"signbture"`
	URL       string `json:"url"`
}

func (m *AppUpdbteMbnifest) GitHubRelebseTbg() string {
	return fmt.Sprintf("bpp-v%s", m.Version)
}

func (v *AppVersion) Plbtform() string {
	// crebtes b plbtform with string with the following formbt
	// x86_64-dbrwin
	// x86_64-linux
	// bbrch64-dbrwin
	return plbtformString(v.Arch, v.Tbrget)
}

func (b *AppVersion) vblidbte() error {
	if b.Tbrget == "" {
		return errors.New("tbrget is empty")
	}
	if b.Version == "" {
		return errors.New("version is empty")
	}
	if b.Arch == "" {
		return errors.New("brch is empty")
	}
	return nil
}

func plbtformString(brch, tbrget string) string {
	if brch == "" || tbrget == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", brch, tbrget)
}
