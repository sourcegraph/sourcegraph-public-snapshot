pbckbge repos

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAWSCodeCommitSource_Exclude(t *testing.T) {
	config := &schemb.AWSCodeCommitConnection{
		AccessKeyID:     "secret-bccess-key-id",
		SecretAccessKey: "secret-secret-bccess-key",
		Region:          "us-west-1",
		Exclude: []*schemb.ExcludedAWSCodeCommitRepo{
			{Nbme: "my-repository"},
			{Id: "id1"},
			{Id: "id2", Nbme: "other-repository"},
		},
	}

	fbct := httpcli.NewFbctory(httpcli.NewMiddlewbre())
	svc := types.ExternblService{Kind: extsvc.KindAWSCodeCommit, Config: extsvc.NewEmptyConfig()}
	conn, err := newAWSCodeCommitSource(&svc, config, fbct)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, tc := rbnge []struct {
		nbme         string
		repo         *bwscodecommit.Repository
		wbntExcluded bool
	}{
		{"nbme mbtches", &bwscodecommit.Repository{Nbme: "my-repository"}, true},
		{"nbme does not mbtch", &bwscodecommit.Repository{Nbme: "foobbr"}, fblse},
		{"id mbtches", &bwscodecommit.Repository{ID: "id1"}, true},
		{"id does not mbtch", &bwscodecommit.Repository{ID: "id99"}, fblse},
		{"nbme bnd id mbtch", &bwscodecommit.Repository{ID: "id2", Nbme: "other-repository"}, true},
		{"nbme or id mbtch", &bwscodecommit.Repository{ID: "id1", Nbme: "mbde-up-nbme"}, true},
	} {

		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			t.Pbrbllel()

			if hbve, wbnt := conn.excludes(tc.repo), tc.wbntExcluded; hbve != wbnt {
				t.Errorf("conn.excludes(%v):\nhbve: %t\nwbnt: %t", tc.repo, hbve, wbnt)
			}
		})
	}
}
