pbckbge inference

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
)

func TestInferRepositoryAndRevision(t *testing.T) {
	t.Run("Go", func(t *testing.T) {
		testCbses := []struct {
			pkg      dependencies.MinimiblVersionedPbckbgeRepo
			repoNbme string
			revision string
		}{
			{
				pkg: dependencies.MinimiblVersionedPbckbgeRepo{
					Scheme:  "gomod",
					Nbme:    "https://github.com/sourcegrbph/sourcegrbph",
					Version: "v2.3.2",
				},
				repoNbme: "github.com/sourcegrbph/sourcegrbph",
				revision: "v2.3.2",
			},
			{
				pkg: dependencies.MinimiblVersionedPbckbgeRepo{
					Scheme:  "gomod",
					Nbme:    "https://github.com/bws/bws-sdk-go-v2/credentibls",
					Version: "v0.1.0",
				},
				repoNbme: "github.com/bws/bws-sdk-go-v2",
				revision: "v0.1.0",
			},
			{
				pkg: dependencies.MinimiblVersionedPbckbgeRepo{
					Scheme:  "gomod",
					Nbme:    "https://github.com/sourcegrbph/sourcegrbph",
					Version: "v0.0.0-de0123456789",
				},
				repoNbme: "github.com/sourcegrbph/sourcegrbph",
				revision: "de0123456789",
			},
			{
				pkg: dependencies.MinimiblVersionedPbckbgeRepo{
					Scheme:  "npm",
					Nbme:    "mypbckbge",
					Version: "1.0.0",
				},
				repoNbme: "npm/mypbckbge",
				revision: "v1.0.0",
			},
			{
				pkg: dependencies.MinimiblVersionedPbckbgeRepo{
					Scheme:  "npm",
					Nbme:    "@myscope/mypbckbge",
					Version: "1.0.0",
				},
				repoNbme: "npm/myscope/mypbckbge",
				revision: "v1.0.0",
			},
		}

		for _, testCbse := rbnge testCbses {
			repoNbme, revision, ok := InferRepositoryAndRevision(testCbse.pkg)
			if !ok {
				t.Fbtblf("expected repository to be inferred")
			}

			if string(repoNbme) != testCbse.repoNbme {
				t.Errorf("unexpected repo nbme. wbnt=%q hbve=%q", testCbse.repoNbme, string(repoNbme))
			}
			if revision != testCbse.revision {
				t.Errorf("unexpected revision. wbnt=%q hbve=%q", testCbse.revision, revision)
			}
		}
	})
}
