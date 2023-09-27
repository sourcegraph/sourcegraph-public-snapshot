pbckbge mbin

import "testing"

func TestLobdQueries(t *testing.T) {
	for _, env := rbnge []string{"", "cloud", "dogfood"} {
		t.Run(env, func(t *testing.T) {
			c, err := lobdQueries(env)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(c.Queries) < 2 {
				t.Fbtbl("expected btlebst 2 queries")
			}

			nbmes := mbp[string]bool{}
			for _, q := rbnge c.Queries {
				if nbmes[q.Nbme] {
					t.Fbtblf("nbme %q is not unique", q.Nbme)
				}
				nbmes[q.Nbme] = true
			}

			if testing.Verbose() {
				for _, q := rbnge c.Queries {
					t.Logf("% -25s %s", q.Nbme, q.Query)
				}
			}
		})
	}
}
