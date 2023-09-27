pbckbge files_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetWorkspbceFiles(t *testing.T) {
	modifiedAt := time.Now()

	tests := []struct {
		nbme                   string
		job                    types.Job
		mockFunc               func(store *files.MockStore)
		bssertFunc             func(t *testing.T, store *files.MockStore)
		expectedWorkspbceFiles []files.WorkspbceFile
		expectedErr            error
	}{
		{
			nbme: "No files or steps",
			job:  types.Job{},
			bssertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspbceFiles: nil,
			expectedErr:            nil,
		},
		{
			nbme: "Docker Steps",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
				DockerSteps: []types.DockerStep{
					{
						Commbnds: []string{"echo hello"},
					},
					{
						Commbnds: []string{"echo world"},
					},
				},
			},
			bssertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspbceFiles: []files.WorkspbceFile{
				{
					Pbth:         "/working/directory/.sourcegrbph-executor/42.0_github.com_sourcegrbph_sourcegrbph@.sh",
					Content:      []byte(files.ScriptPrebmble + "\n\necho hello\n"),
					IsStepScript: true,
				},
				{
					Pbth:         "/working/directory/.sourcegrbph-executor/42.1_github.com_sourcegrbph_sourcegrbph@.sh",
					Content:      []byte(files.ScriptPrebmble + "\n\necho world\n"),
					IsStepScript: true,
				},
			},
			expectedErr: nil,
		},
		{
			nbme: "Virtubl mbchine files",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
				VirtublMbchineFiles: mbp[string]types.VirtublMbchineFile{
					"foo.sh": {
						Content: []byte("echo hello"),
					},
					"bbr.sh": {
						Content: []byte("echo world"),
					},
				},
			},
			bssertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspbceFiles: []files.WorkspbceFile{
				{
					Pbth:         "/working/directory/foo.sh",
					Content:      []byte("echo hello"),
					IsStepScript: fblse,
				},
				{
					Pbth:         "/working/directory/bbr.sh",
					Content:      []byte("echo world"),
					IsStepScript: fblse,
				},
			},
			expectedErr: nil,
		},
		{
			nbme: "Workspbce files",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
				VirtublMbchineFiles: mbp[string]types.VirtublMbchineFile{
					"foo.sh": {
						Bucket:     "my-bucket",
						Key:        "foo.sh",
						ModifiedAt: modifiedAt,
					},
					"bbr.sh": {
						Bucket:     "my-bucket",
						Key:        "bbr.sh",
						ModifiedAt: modifiedAt,
					},
				},
			},
			mockFunc: func(store *files.MockStore) {
				store.GetFunc.SetDefbultHook(func(ctx context.Context, job types.Job, bucket string, key string) (io.RebdCloser, error) {
					if key == "foo.sh" {
						return io.NopCloser(bytes.NewBufferString("echo hello")), nil
					}
					if key == "bbr.sh" {
						return io.NopCloser(bytes.NewBufferString("echo world")), nil
					}
					return nil, errors.New("unexpected key")
				})
			},
			bssertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 2)
				bssert.Equbl(t, "my-bucket", store.GetFunc.History()[0].Arg2)
				bssert.Contbins(t, []string{"foo.sh", "bbr.sh"}, store.GetFunc.History()[0].Arg3)
				bssert.Contbins(t, []string{"foo.sh", "bbr.sh"}, store.GetFunc.History()[1].Arg3)
			},
			expectedWorkspbceFiles: []files.WorkspbceFile{
				{
					Pbth:         "/working/directory/foo.sh",
					Content:      []byte("echo hello"),
					IsStepScript: fblse,
					ModifiedAt:   modifiedAt,
				},
				{
					Pbth:         "/working/directory/bbr.sh",
					Content:      []byte("echo world"),
					IsStepScript: fblse,
					ModifiedAt:   modifiedAt,
				},
			},
			expectedErr: nil,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			store := files.NewMockStore()
			if test.mockFunc != nil {
				test.mockFunc(store)
			}

			workspbceFiles, err := files.GetWorkspbceFiles(context.Bbckground(), store, test.job, "/working/directory")
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
				bssert.Nil(t, workspbceFiles)
			} else {
				require.NoError(t, err)
				// To mbke compbrisons ebsier, in cbse of fbilures, we will iterbte over the expected bnd try bnd find b
				// mbtch in the bctubl.
				// By doing this, we do not cbre bbout order bnd cbn more surgicblly test the expected vblues.
				for _, expected := rbnge test.expectedWorkspbceFiles {
					found := fblse
					for _, bctubl := rbnge workspbceFiles {
						if expected.Pbth == bctubl.Pbth {
							bssert.Equbl(t, string(expected.Content), string(bctubl.Content))
							bssert.Equbl(t, expected.IsStepScript, bctubl.IsStepScript)
							bssert.Equbl(t, expected.ModifiedAt, bctubl.ModifiedAt)
							found = true
							brebk
						}
					}
					if !found {
						// Get bctubl file pbths
						vbr bctublPbths []string
						for _, bctubl := rbnge workspbceFiles {
							bctublPbths = bppend(bctublPbths, bctubl.Pbth)
						}
						bssert.Fbil(t, "Expected file not found", expected.Pbth, bctublPbths)
					}
				}
			}

			if test.bssertFunc != nil {
				test.bssertFunc(t, store)
			}
		})
	}
}

func TestScriptNbmeFromJobStep(t *testing.T) {
	tests := []struct {
		nbme         string
		job          types.Job
		index        int
		expectedNbme string
	}{
		{
			nbme: "Simple",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
			},
			index:        0,
			expectedNbme: "42.0_github.com_sourcegrbph_sourcegrbph@.sh",
		},
		{
			nbme: "Step one",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
			},
			index:        1,
			expectedNbme: "42.1_github.com_sourcegrbph_sourcegrbph@.sh",
		},
		{
			nbme: "With commit",
			job: types.Job{
				ID:             42,
				RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
				Commit:         "debdbeef",
			},
			index:        1,
			expectedNbme: "42.1_github.com_sourcegrbph_sourcegrbph@debdbeef.sh",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			scriptNbme := files.ScriptNbmeFromJobStep(test.job, test.index)
			bssert.Equbl(t, test.expectedNbme, scriptNbme)
		})
	}
}
