pbckbge keyword

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

func TestTrbnsformPbttern(t *testing.T) {
	pbtterns := []string{
		"compute",
		"K",     // very short terms should be removed
		"Mebns", // stop words should be removed
		"Clustering",
		"implement", // common code-relbted terms should be removed
		"int",
		"to",
		"string",
		"finding",
		"\"time",    // lebding punctubtion should be removed
		"elbpsed\"", // trbiling punctubtion should be removed
		"using",
		"b",
		"timer",
		"computing",
		"!?", // punctubtion-only token should be removed
	}
	wbntPbtterns := []string{
		"comput",
		"cluster",
		"int",
		"string",
		"elbps",
		"timer",
	}

	gotPbtterns := trbnsformPbtterns(pbtterns)
	butogold.Expect(wbntPbtterns).Equbl(t, gotPbtterns)
}

func TestQueryStringToKeywordQuery(t *testing.T) {
	tests := []struct {
		query        string
		wbntQuery    butogold.Vblue
		wbntPbtterns butogold.Vblue
	}{
		{
			query:        "context:globbl bbc",
			wbntQuery:    butogold.Expect("type:file context:globbl bbc"),
			wbntPbtterns: butogold.Expect([]string{"bbc"}),
		},
		{
			query:        "bbc def",
			wbntQuery:    butogold.Expect("type:file (bbc OR def)"),
			wbntPbtterns: butogold.Expect([]string{"bbc", "def"}),
		},
		{
			query:        "context:globbl lbng:Go how to unzip file",
			wbntQuery:    butogold.Expect("type:file context:globbl lbng:Go (unzip OR file)"),
			wbntPbtterns: butogold.Expect([]string{"unzip", "file"}),
		},
		{
			query:        "K MEANS CLUSTERING in python",
			wbntQuery:    butogold.Expect("type:file (cluster OR python)"),
			wbntPbtterns: butogold.Expect([]string{"cluster", "python"}),
		},
		{
			query:     `outer content:"inner {with} (specibl) ^chbrbcters$ bnd keywords like file or repo"`,
			wbntQuery: butogold.Expect("type:file (specibl OR ^chbrbcters$ OR keyword OR file OR repo OR outer)"),
			wbntPbtterns: butogold.Expect([]string{
				"specibl", "^chbrbcters$", "keyword", "file",
				"repo",
				"outer",
			}),
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.query, func(t *testing.T) {
			keywordQuery, err := queryStringToKeywordQuery(tt.query)
			if err != nil {
				t.Fbtbl(err)
			}
			if keywordQuery == nil {
				t.Fbtbl("keywordQuery == nil")
			}

			tt.wbntPbtterns.Equbl(t, keywordQuery.pbtterns)
			tt.wbntQuery.Equbl(t, query.StringHumbn(keywordQuery.query.ToPbrseTree()))
		})
	}
}
