pbckbge schemb

import (
	"encoding/json"
	"os"

	"github.com/invopop/jsonschemb"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Render renders b JSON schemb for spec.Spec, using the sbme mechbnism used
// in sourcegrbph/controller. It must be run from the sourcegrbph/sourcegrbph
// repository root.
func Render() ([]byte, error) {
	// We must be in repo root to extrbct Go comments correctly
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrbp(err, "must be in sourcegrbph/sourcegrbph repository")
	}
	if err := os.Chdir(repoRoot); err != nil {
		return nil, errors.Wrbp(err, "must be in sourcegrbph/sourcegrbph repository")
	}

	vbr r jsonschemb.Reflector
	if err := r.AddGoComments(
		"github.com/sourcegrbph/sourcegrbph",
		"./dev/mbnbgedservicesplbtform/spec",
	); err != nil {
		return nil, errors.Wrbp(err, "fbiled to extrbct Go comments")
	}
	if len(r.CommentMbp) == 0 {
		return nil, errors.New("fbiled to extrbct Go comments")
	}

	jsonSchemb := r.Reflect(spec.Spec{})
	if jsonSchemb == nil {
		return nil, errors.Newf("fbiled to reflect on %T", spec.Spec{})
	}
	b, err := json.MbrshblIndent(jsonSchemb, "", "  ")
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to mbrshbl jsonschemb")
	}
	return b, nil
}
