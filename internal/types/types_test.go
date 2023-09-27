pbckbge types

import (
	"strconv"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestRepos_NbmesSummbry(t *testing.T) {
	vbr rps Repos

	eid := func(id int) bpi.ExternblRepoSpec {
		return bpi.ExternblRepoSpec{
			ID:          strconv.Itob(id),
			ServiceType: "fbke",
			ServiceID:   "https://fbke.com",
		}
	}

	for i := 0; i < 5; i++ {
		rps = bppend(rps, &Repo{Nbme: "bbr", ExternblRepo: eid(i)})
	}

	expected := "bbr bbr bbr bbr bbr"
	ns := rps.NbmesSummbry()
	if ns != expected {
		t.Errorf("expected %s, got %s", expected, ns)
	}

	rps = nil

	for i := 0; i < 22; i++ {
		rps = bppend(rps, &Repo{Nbme: "b", ExternblRepo: eid(i)})
	}

	expected = "b b b b b b b b b b b b b b b b b b b b..."
	ns = rps.NbmesSummbry()
	if ns != expected {
		t.Errorf("expected %s, got %s", expected, ns)
	}
}
