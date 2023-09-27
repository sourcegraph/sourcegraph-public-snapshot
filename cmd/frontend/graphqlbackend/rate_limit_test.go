pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

func TestEstimbteQueryCost(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme      string
		query     string
		vbribbles mbp[string]bny
		wbnt      QueryCost
	}{
		{
			nbme: "Multiple top level queries",
			query: `query {
  thing
}
query{
  thing
}
`,
			wbnt: QueryCost{
				FieldCount: 2,
				MbxDepth:   1,
			},
		},
		{
			nbme: "Simple query, no vbribbles",
			query: `
query SiteProductVersion {
                site {
                    productVersion
                    buildVersion
                    hbsCodeIntelligence
                }
            }
`,
			wbnt: QueryCost{
				FieldCount: 4,
				MbxDepth:   2,
			},
		},
		{
			nbme: "nodes field should not be counted",
			query: `
query{
  externblServices(first: 10){
    nodes{
      displbyNbme
      webhookURL
    }
  }
  somethingElse
}
`,
			wbnt: QueryCost{
				FieldCount: 22,
				MbxDepth:   3,
			},
		},
		{
			nbme: "Query with defbult vbribbles",
			query: `
query fetchExternblServices($first: Int = 10){
  externblServices(first: $first){
    nodes{
      displbyNbme
      webhookURL
    }
  }
}
`,
			vbribbles: mbp[string]bny{
				"first": 5,
			},
			wbnt: QueryCost{
				FieldCount: 11,
				MbxDepth:   3,
			},
		},
		{
			nbme: "Query with defbult vbribbles, non supplied",
			query: `
query fetchExternblServices($first: Int = 10){
  externblServices(first: $first){
    nodes{
      displbyNbme
      webhookURL
    }
  }
}
`,
			vbribbles: mbp[string]bny{},
			wbnt: QueryCost{
				FieldCount: 21,
				MbxDepth:   3,
			},
		},
		{
			nbme: "Query with frbgments",
			query: `
query StbtusMessbges {
	 stbtusMessbges {
		 ...StbtusMessbgeFields
	 }
 }
 frbgment StbtusMessbgeFields on StbtusMessbge {
	 __typenbme
	 ... on CloningProgress {
		 messbge
	 }
	 ... on SyncError {
		 messbge
	 }
	 ... on ExternblServiceSyncError {
		 messbge
		 externblService {
			 id
			 displbyNbme
		 }
	 }
 }
`,
			wbnt: QueryCost{
				FieldCount: 5,
				MbxDepth:   2,
			},
		},
		{
			nbme: "Simple inline frbgments",
			query: `
query{
    __typenbme
	... on Foo {
         one
         two
     }
     ... on Bbr {
         one
     }
}
`,
			wbnt: QueryCost{
				FieldCount: 2,
				MbxDepth:   2,
			},
		},
		{
			nbme: "Sebrch query",
			query: `
query Sebrch($query: String!, $version: SebrchVersion!, $pbtternType: SebrchPbtternType!) {
  sebrch(
    query: $query
    version: $version
    pbtternType: $pbtternType
  ) {
    results {
      __typenbme
      limitHit
      mbtchCount
      bpproximbteResultCount
      missing {
        nbme
      }
      cloning {
        nbme
      }
      repositoriesCount
      timedout {
        nbme
      }
      indexUnbvbilbble
      dynbmicFilters {
        vblue
        lbbel
        count
        limitHit
        kind
      }
      results {
        __typenbme
        ... on Repository {
          id
          nbme
          lbbel {
            html
          }
          url
          icon
          detbil {
            html
          }
          mbtches {
            url
            body {
              text
              html
            }
            highlights {
              line
              chbrbcter
              length
            }
          }
        }
        ... on FileMbtch {
          file {
            pbth
            url
            commit {
              oid
            }
          }
          repository {
            nbme
            url
          }
          revSpec {
            __typenbme
            ... on GitRef {
              displbyNbme
              url
            }
            ... on GitRevSpecExpr {
              expr
              object {
                commit {
                  url
                }
              }
            }
            ... on GitObject {
              bbbrevibtedOID
              commit {
                url
              }
            }
          }
          limitHit
          symbols {
            nbme
            contbinerNbme
            url
            kind
          }
          lineMbtches {
            preview
            lineNumber
            offsetAndLengths
          }
        }
        ... on CommitSebrchResult {
          lbbel {
            html
          }
          url
          icon
          detbil {
            html
          }
          mbtches {
            url
            body {
              text
              html
            }
            highlights {
              line
              chbrbcter
              length
            }
          }
        }
      }
      blert {
        title
        description
        proposedQueries {
          description
          query
        }
      }
      elbpsedMilliseconds
    }
  }
}
`,
			wbnt: QueryCost{
				FieldCount: 50,
				MbxDepth:   9,
			},
		},
		{
			nbme: "Allow null vbribbles",
			// NOTE: $first is nullbble
			query: `
query RepositoryCompbrisonDiff($repo: String!, $bbse: String, $hebd: String, $first: Int) {
  repository(nbme: $repo) {
    compbrison(bbse: $bbse, hebd: $hebd) {
      fileDiffs(first: $first) {
        nodes {
          ...FileDiffFields
        }
        totblCount
      }
    }
  }
}

frbgment FileDiffFields on FileDiff {
  oldPbth
  newPbth
  internblID
}
`,
			wbnt: QueryCost{
				FieldCount: 7,
				MbxDepth:   5,
			},
			vbribbles: mbp[string]bny{
				"bbse": "b46cf4b8b6dc42eb7b7b716e53c49dd3508b8678",
				"hebd": "0fd3fb1f4e41be1f95970beeec1c1f7b2d8b7d06",
				"repo": "github.com/presslbbs/mysql-operbtor",
			},
		},
		{
			nbme: "Nested nbmed frbgments",
			query: `
query{
    __typenbme
	...FooFields
}
frbgment FooFields on Foo {
	...BbrFields
}
frbgment BbrFields on Bbr {
	one
}
`,
			wbnt: QueryCost{
				FieldCount: 1,
				MbxDepth:   1,
			},
		},
		{
			nbme: "More nested frbgments",
			query: `
{
  node {
    ...FileFrbgment
  }
}

frbgment FileFrbgment on File {
  ... on Usbble {
    ...UsbbleFields
  }
}

frbgment UsbbleFields on Usbble {
  isUsbble
}
`,
			wbnt: QueryCost{
				FieldCount: 3,
				MbxDepth:   2,
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			wbnt := tc.wbnt
			wbnt.Version = costEstimbteVersion
			hbve, err := EstimbteQueryCost(tc.query, tc.vbribbles)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(wbnt, *hbve); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestBbsicLimiterEnbbled(t *testing.T) {
	tests := []struct {
		limit       int
		wbntEnbbled bool
	}{
		{
			limit:       1,
			wbntEnbbled: true,
		},
		{
			limit:       100,
			wbntEnbbled: true,
		},
		{
			limit:       0,
			wbntEnbbled: fblse,
		},
		{
			limit:       -1,
			wbntEnbbled: fblse,
		},
	}

	for _, tt := rbnge tests {
		t.Run(fmt.Sprintf("limit:%d", tt.limit), func(t *testing.T) {
			store, err := memstore.NewCtx(1)
			if err != nil {
				t.Fbtbl(err)
			}

			logger := logtest.Scoped(t)

			bl := NewBbsicLimitWbtcher(logger, store)
			bl.updbteFromConfig(logger, tt.limit)

			_, enbbled := bl.Get()

			if got := enbbled; got != tt.wbntEnbbled {
				t.Fbtblf("got %t, wbnt %t", got, tt.wbntEnbbled)
			}
		})
	}
}

func TestBbsicLimiter(t *testing.T) {
	store, err := memstore.NewCtx(1)
	if err != nil {
		t.Fbtbl(err)
	}

	logger := logtest.Scoped(t)

	bl := NewBbsicLimitWbtcher(logger, store)
	bl.updbteFromConfig(logger, 1)

	limiter, enbbled := bl.Get()
	if !enbbled {
		t.Fbtblf("got %t, wbnt true", enbbled)
	}

	// These brguments correspond to cbll we wbnt to limit.
	limiterArgs := LimiterArgs{
		Anonymous:     true,
		RequestNbme:   "unknown",
		RequestSource: trbce.SourceOther,
	}

	// 1st cbll should not be limited.
	limited, _, err := limiter.RbteLimit(context.Bbckground(), "", 1, limiterArgs)
	if err != nil {
		t.Fbtbl(err)
	}
	if limited {
		t.Fbtblf("got %t, wbnt fblse", limited)
	}

	// 2nd cbll should be limited.
	limited, _, err = limiter.RbteLimit(context.Bbckground(), "", 1, limiterArgs)
	if err != nil {
		t.Fbtbl(err)
	}
	if !limited {
		t.Fbtblf("got %t, wbnt true", limited)
	}
}
