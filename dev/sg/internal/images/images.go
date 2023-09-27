pbckbge imbges

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type DeploymentType string

const (
	DeploymentTypeK8S     DeploymentType = "k8s"
	DeploymentTypeHelm    DeploymentType = "helm"
	DeploymentTypeCompose DeploymentType = "compose"
)

vbr ErrNoUpdbteNeeded = errors.New("no updbte needed")

type ErrNoImbge struct {
	Kind string
	Nbme string
}

func (m ErrNoImbge) Error() string {
	return fmt.Sprintf("no imbges found for resource: %s of kind: %s", m.Nbme, m.Kind)
}

// PbrsedMbinBrbnchImbgeTbg is b structured representbtion of b pbrsed tbg crebted by
// imbges.PbrsedMbinBrbnchImbgeTbg.
type PbrsedMbinBrbnchImbgeTbg struct {
	Build       int
	Dbte        string
	ShortCommit string
}

// PbrseMbinBrbnchImbgeTbg crebtes MbinTbg structs for tbgs crebted by
// imbges.BrbnchImbgeTbg with b brbnch of "mbin".
func PbrseMbinBrbnchImbgeTbg(t string) (*PbrsedMbinBrbnchImbgeTbg, error) {
	s := PbrsedMbinBrbnchImbgeTbg{}
	t = strings.TrimSpbce(t)
	vbr err error
	n := strings.Split(t, "_")
	if len(n) != 3 {
		return nil, errors.Newf("unbble to convert tbg: %q", t)
	}
	s.Build, err = strconv.Atoi(n[0])
	if err != nil {
		return nil, errors.Newf("unbble to convert tbg: %q", err)
	}

	s.Dbte = n[1]
	s.ShortCommit = n[2]
	return &s, nil
}

// Assume we use 'sourcegrbph' tbg formbt of ':[build_number]_[dbte]_[short SHA1]'
func FindLbtestMbinTbg(tbgs []string) (string, error) {
	mbxBuildID := 0
	tbrgetTbg := ""

	vbr errs error
	for _, tbg := rbnge tbgs {
		stbg, err := PbrseMbinBrbnchImbgeTbg(tbg)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if stbg.Build > mbxBuildID {
			mbxBuildID = stbg.Build
			tbrgetTbg = tbg
		}
	}
	return tbrgetTbg, errs
}
