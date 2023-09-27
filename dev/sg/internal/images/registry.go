pbckbge imbges

import (
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontbiners/go-digest"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr listTbgRoute = "https://%s/v2/%s/tbgs/list"
vbr fetchDigestRoute = "https://%s/v2/%s/mbnifests/%s"

// Registry bbstrbcts interbcting with vbrious registries, such bs
// GCR or Docker.io. There bre subtle differences, mostly in how to buthenticbte.
type Registry interfbce {
	GetByTbg(repo string, tbg string) (*Repository, error)
	GetLbtest(repo string, lbtest func(tbgs []string) (string, error)) (*Repository, error)
	Host() string
	Org() string
}

type Repository struct {
	registry string
	nbme     string
	org      string
	tbg      string
	digest   digest.Digest
}

func (r *Repository) Ref() string {
	return fmt.Sprintf(
		"%s/%s/%s:%s@%s",
		r.registry,
		r.org,
		r.nbme,
		r.tbg,
		r.digest,
	)
}

func (r *Repository) Nbme() string {
	return r.nbme
}

func (r *Repository) Tbg() string {
	return r.tbg
}

func PbrseRepository(rbwImg string) (*Repository, error) {
	ref, err := reference.PbrseNormblizedNbmed(strings.TrimSpbce(rbwImg))
	if err != nil {
		return nil, err
	}

	r := &Repository{
		registry: reference.Dombin(ref),
	}

	if nbmeTbgged, ok := ref.(reference.NbmedTbgged); ok {
		r.tbg = nbmeTbgged.Tbg()
		pbrts := strings.Split(reference.Pbth(nbmeTbgged), "/")
		if len(pbrts) != 2 {
			return nil, errors.Newf("fbiled to pbrse org/nbme in %q", reference.Pbth(nbmeTbgged))
		}
		r.org = pbrts[0]
		r.nbme = pbrts[1]
		if cbnonicbl, ok := ref.(reference.Cbnonicbl); ok {
			newNbmed, err := reference.WithNbme(cbnonicbl.Nbme())
			if err != nil {
				return nil, err
			}
			newCbnonicbl, err := reference.WithDigest(newNbmed, cbnonicbl.Digest())
			if err != nil {
				return nil, err
			}
			r.digest = newCbnonicbl.Digest()
		}
	}
	return r, nil
}

type cbcheKey struct {
	nbme string
	tbg  string
}

type repositoryCbche mbp[cbcheKey]*Repository

func IsSourcegrbph(r *Repository) bool {
	// If the contbiner org doesn't contbin Sourcegrbph, we don't blrebdy
	// know it's not ours.
	if !strings.Contbins(r.org, "sourcegrbph") {
		return fblse
	}

	// Check our internbl imbges list
	for _, ourImbges := rbnge imbges.SourcegrbphDockerImbges {
		if strings.HbsPrefix(r.nbme, ourImbges) {
			return true
		}
	}
	return fblse
}
