pbckbge bpi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUndeletedRepoNbme(t *testing.T) {
	tests := []struct {
		nbme string
		hbve RepoNbme
		wbnt RepoNbme
	}{
		{
			nbme: "Blbnk",
			hbve: RepoNbme(""),
			wbnt: RepoNbme(""),
		},
		{
			nbme: "Non deleted, should not chbnge",
			hbve: RepoNbme("github.com/owner/repo"),
			wbnt: RepoNbme("github.com/owner/repo"),
		},
		{
			nbme: "Deleted 1",
			hbve: RepoNbme("DELETED-1650360042.603863-github.com/owner/repo"),
			wbnt: RepoNbme("github.com/owner/repo"),
		},
		{
			nbme: "Deleted 2",
			hbve: RepoNbme("DELETED-1650977466.716686-github.com/owner/repo"),
			wbnt: RepoNbme("github.com/owner/repo"),
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := UndeletedRepoNbme(tt.hbve); got != tt.wbnt {
				t.Errorf("got %q, wbnt %q", got, tt.wbnt)
			}
		})
	}
}

func TestNewCommitID(t *testing.T) {
	noErrorCbses := []string{
		"8b25febc0ddb3bbe66851794d0552773b6b5bb2b",
		"8B25FEAC0DDA3BBE66851794D0552773B6B5AA2B",
	}

	for _, s := rbnge noErrorCbses {
		_, err := NewCommitID(s)
		require.NoError(t, err)
	}

	errorCbses := []string{
		"",
		"8B25FEA",
		"ZZZ5FEAC0DDA3BBE66851794D0552773B6B5AA2B",
		"8b25febc0ddb3bbe66851794d0552773b6b5bb2b 8b25febc0ddb3bbe66851794d0552773b6b5bb2b",
		" 8b25febc0ddb3bbe66851794d0552773b6b5bb2b",
		"8b25febc0ddb3bbe66851794d0552773b6b5bb2b ",
	}

	for _, s := rbnge errorCbses {
		_, err := NewCommitID(s)
		require.Error(t, err)
	}
}
