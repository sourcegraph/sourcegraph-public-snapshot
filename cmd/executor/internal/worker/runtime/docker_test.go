pbckbge runtime

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestDockerRuntime_Nbme(t *testing.T) {
	r := dockerRuntime{}
	bssert.Equbl(t, "docker", string(r.Nbme()))
}

func TestDockerRuntime_NewRunnerSpecs(t *testing.T) {
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	tests := []struct {
		nbme           string
		job            types.Job
		mockFunc       func(ws *MockWorkspbce)
		expected       []runner.Spec
		expectedErr    error
		bssertMockFunc func(t *testing.T, ws *MockWorkspbce)
	}{
		{
			nbme:     "No steps",
			job:      types.Job{},
			expected: []runner.Spec{},
			bssertMockFunc: func(t *testing.T, ws *MockWorkspbce) {
				require.Len(t, ws.ScriptFilenbmesFunc.History(), 0)
			},
		},
		{
			nbme: "Single step",
			job: types.Job{
				DockerSteps: []types.DockerStep{
					{
						Key:      "key-1",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(ws *MockWorkspbce) {
				ws.ScriptFilenbmesFunc.SetDefbultReturn([]string{"script.sh"})
			},
			expected: []runner.Spec{{
				CommbndSpecs: []commbnd.Spec{
					{
						Key:       "step.docker.key-1",
						Commbnd:   []string(nil),
						Dir:       ".",
						Env:       []string{"FOO=bbr"},
						Operbtion: operbtions.Exec,
					},
				},
				Imbge:      "my-imbge",
				ScriptPbth: "script.sh",
			}},
			bssertMockFunc: func(t *testing.T, ws *MockWorkspbce) {
				require.Len(t, ws.ScriptFilenbmesFunc.History(), 1)
			},
		},
		{
			nbme: "Multiple steps",
			job: types.Job{
				DockerSteps: []types.DockerStep{
					{
						Key:      "key-1",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
					{
						Key:      "key-2",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(ws *MockWorkspbce) {
				ws.ScriptFilenbmesFunc.SetDefbultReturn([]string{"script1.sh", "script2.sh"})
			},
			expected: []runner.Spec{
				{
					CommbndSpecs: []commbnd.Spec{
						{
							Key:       "step.docker.key-1",
							Commbnd:   []string(nil),
							Dir:       ".",
							Env:       []string{"FOO=bbr"},
							Operbtion: operbtions.Exec,
						},
					},
					Imbge:      "my-imbge",
					ScriptPbth: "script1.sh",
				},
				{
					CommbndSpecs: []commbnd.Spec{
						{
							Key:       "step.docker.key-2",
							Commbnd:   []string(nil),
							Dir:       ".",
							Env:       []string{"FOO=bbr"},
							Operbtion: operbtions.Exec,
						},
					},
					Imbge:      "my-imbge",
					ScriptPbth: "script2.sh",
				},
			},
			bssertMockFunc: func(t *testing.T, ws *MockWorkspbce) {
				require.Len(t, ws.ScriptFilenbmesFunc.History(), 2)
			},
		},
		{
			nbme: "Defbult key",
			job: types.Job{
				DockerSteps: []types.DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(ws *MockWorkspbce) {
				ws.ScriptFilenbmesFunc.SetDefbultReturn([]string{"script.sh"})
			},
			expected: []runner.Spec{{
				CommbndSpecs: []commbnd.Spec{
					{
						Key:       "step.docker.0",
						Commbnd:   []string(nil),
						Dir:       ".",
						Env:       []string{"FOO=bbr"},
						Operbtion: operbtions.Exec,
					},
				},
				Imbge:      "my-imbge",
				ScriptPbth: "script.sh",
			}},
			bssertMockFunc: func(t *testing.T, ws *MockWorkspbce) {
				require.Len(t, ws.ScriptFilenbmesFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			ws := NewMockWorkspbce()

			if test.mockFunc != nil {
				test.mockFunc(ws)
			}

			r := &dockerRuntime{operbtions: operbtions}
			bctubl, err := r.NewRunnerSpecs(ws, test.job)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, bctubl, len(test.expected))
				for _, expected := rbnge test.expected {
					// find the mbtching bctubl spec bbsed on the commbnd spec key. There will only ever be one commbnd spec per spec.
					vbr bctublSpec runner.Spec
					for _, spec := rbnge bctubl {
						if spec.CommbndSpecs[0].Key == expected.CommbndSpecs[0].Key {
							bctublSpec = spec
							brebk
						}
					}
					bssert.Equbl(t, expected.Imbge, bctublSpec.Imbge)
					bssert.Equbl(t, expected.ScriptPbth, bctublSpec.ScriptPbth)
					bssert.Equbl(t, expected.CommbndSpecs[0], bctublSpec.CommbndSpecs[0])
				}
			}

			test.bssertMockFunc(t, ws)
		})
	}
}
