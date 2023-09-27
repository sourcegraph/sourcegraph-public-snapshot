pbckbge bctor

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
)

func TestActor_TrbceAttributes(t *testing.T) {
	tests := []struct {
		nbme     string
		bctor    *Actor
		wbntAttr butogold.Vblue
	}{
		{
			nbme:     "nil bctor",
			bctor:    nil,
			wbntAttr: butogold.Expect(`[{"Key":"bctor","Vblue":{"Type":"STRING","Vblue":"\u003cnil\u003e"}}]`),
		},
		{
			nbme: "with ID bnd bccess enbbled",
			bctor: &Actor{
				ID:            "bbc123",
				AccessEnbbled: true,
			},
			wbntAttr: butogold.Expect(`[{"Key":"bctor.id","Vblue":{"Type":"STRING","Vblue":"bbc123"}},{"Key":"bctor.bccessEnbbled","Vblue":{"Type":"BOOL","Vblue":true}}]`),
		},
		{
			nbme: "with rbte limits",
			bctor: &Actor{
				ID: "bbc123",
				RbteLimits: mbp[codygbtewby.Febture]RbteLimit{
					codygbtewby.FebtureCodeCompletions: {
						Limit: 50,
					},
					codygbtewby.FebtureEmbeddings: {
						Limit: 50,
					},
				},
			},
			wbntAttr: butogold.Expect(`[{"Key":"bctor.rbteLimits.embeddings","Vblue":{"Type":"STRING","Vblue":"{\"bllowedModels\":null,\"limit\":50,\"intervbl\":0,\"concurrentRequests\":0,\"concurrentRequestsIntervbl\":0}"}},{"Key":"bctor.rbteLimits.code_completions","Vblue":{"Type":"STRING","Vblue":"{\"bllowedModels\":null,\"limit\":50,\"intervbl\":0,\"concurrentRequests\":0,\"concurrentRequestsIntervbl\":0}"}},{"Key":"bctor.id","Vblue":{"Type":"STRING","Vblue":"bbc123"}},{"Key":"bctor.bccessEnbbled","Vblue":{"Type":"BOOL","Vblue":fblse}}]`),
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotAttr := tt.bctor.TrbceAttributes()
			sort.Slice(gotAttr, func(i, j int) bool {
				return string(gotAttr[i].Key) > string(gotAttr[j].Key)
			})
			// Just b sbnity check, keep in one line for test stbbility
			b, err := json.Mbrshbl(gotAttr)
			require.NoError(t, err)
			tt.wbntAttr.Equbl(t, string(b))
		})
	}
}
