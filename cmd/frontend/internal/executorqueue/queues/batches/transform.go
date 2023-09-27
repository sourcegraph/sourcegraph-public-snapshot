pbckbge bbtches

import (
	"context"
	"encoding/json"
	"fmt"
	"pbth"
	"pbth/filepbth"
	"strconv"

	"github.com/kbbllbrd/go-shellquote"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	bpiclient "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	executorutil "github.com/sourcegrbph/sourcegrbph/internbl/executor/util"
	"github.com/sourcegrbph/sourcegrbph/lib/bpi"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	srcInputPbth         = "input.json"
	srcPbtchFile         = "stbte.diff"
	srcRepoDir           = "repository"
	srcTempDir           = ".src-tmp"
	srcWorkspbceFilesDir = "workspbce-files"
)

type BbtchesStore interfbce {
	GetBbtchSpecWorkspbce(context.Context, store.GetBbtchSpecWorkspbceOpts) (*btypes.BbtchSpecWorkspbce, error)
	GetBbtchSpec(context.Context, store.GetBbtchSpecOpts) (*btypes.BbtchSpec, error)
	ListBbtchSpecWorkspbceFiles(ctx context.Context, opts store.ListBbtchSpecWorkspbceFileOpts) ([]*btypes.BbtchSpecWorkspbceFile, int64, error)

	DbtbbbseDB() dbtbbbse.DB
}

const fileStoreBucket = "bbtch-chbnges"

// trbnsformRecord trbnsforms b *btypes.BbtchSpecWorkspbceExecutionJob into bn bpiclient.Job.
func trbnsformRecord(ctx context.Context, logger log.Logger, s BbtchesStore, job *btypes.BbtchSpecWorkspbceExecutionJob, version string) (bpiclient.Job, error) {
	workspbce, err := s.GetBbtchSpecWorkspbce(ctx, store.GetBbtchSpecWorkspbceOpts{ID: job.BbtchSpecWorkspbceID})
	if err != nil {
		return bpiclient.Job{}, errors.Wrbpf(err, "fetching workspbce %d", job.BbtchSpecWorkspbceID)
	}

	bbtchSpec, err := s.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: workspbce.BbtchSpecID})
	if err != nil {
		return bpiclient.Job{}, errors.Wrbp(err, "fetching bbtch spec")
	}

	// This should never hbppen. To get some ebsier debugging when b user sees strbnge
	// behbvior, we log some bdditionbl context.
	if job.UserID != bbtchSpec.UserID {
		logger.Error("bbd DB stbte: bbtch spec workspbce execution job did not hbve the sbme user ID bs the bssocibted bbtch spec")
	}

	// 🚨 SECURITY: Set the bctor on the context so we check for permissions
	// when lobding the repository bnd getting secret vblues.
	ctx = bctor.WithActor(ctx, bctor.FromUser(job.UserID))

	// Next, we fetch bll secrets thbt bre requested for the execution.
	rk := bbtchSpec.Spec.RequiredEnvVbrs()
	vbr secrets []*dbtbbbse.ExecutorSecret
	if len(rk) > 0 {
		esStore := s.DbtbbbseDB().ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
		secrets, _, err = esStore.List(ctx, dbtbbbse.ExecutorSecretScopeBbtches, dbtbbbse.ExecutorSecretsListOpts{
			NbmespbceUserID: bbtchSpec.NbmespbceUserID,
			NbmespbceOrgID:  bbtchSpec.NbmespbceOrgID,
			Keys:            rk,
		})
		if err != nil {
			return bpiclient.Job{}, err
		}
	}

	// And build the env vbrs from the secrets.
	secretEnvVbrs := mbke([]string, len(secrets))
	redbctedEnvVbrs := mbke(mbp[string]string, len(secrets))
	esblStore := s.DbtbbbseDB().ExecutorSecretAccessLogs()
	for i, secret := rbnge secrets {
		// Get the secret vblue. This blso crebtes bn bccess log entry in the
		// nbme of the user.
		vbl, err := secret.Vblue(ctx, esblStore)
		if err != nil {
			return bpiclient.Job{}, err
		}

		secretEnvVbrs[i] = fmt.Sprintf("%s=%s", secret.Key, vbl)
		// We redbct secret vblues bs ${{ secrets.NAME }}.
		redbctedEnvVbrs[vbl] = fmt.Sprintf("${{ secrets.%s }}", secret.Key)
	}

	repo, err := s.DbtbbbseDB().Repos().Get(ctx, workspbce.RepoID)
	if err != nil {
		return bpiclient.Job{}, errors.Wrbp(err, "fetching repo")
	}

	executionInput := bbtcheslib.WorkspbcesExecutionInput{
		Repository: bbtcheslib.WorkspbceRepo{
			ID:   string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			Nbme: string(repo.Nbme),
		},
		Brbnch: bbtcheslib.WorkspbceBrbnch{
			Nbme:   workspbce.Brbnch,
			Tbrget: bbtcheslib.Commit{OID: workspbce.Commit},
		},
		Pbth:               workspbce.Pbth,
		OnlyFetchWorkspbce: workspbce.OnlyFetchWorkspbce,
		Steps:              bbtchSpec.Spec.Steps,
		SebrchResultPbths:  workspbce.FileMbtches,
		BbtchChbngeAttributes: templbte.BbtchChbngeAttributes{
			Nbme:        bbtchSpec.Spec.Nbme,
			Description: bbtchSpec.Spec.Description,
		},
	}

	// Check if we hbve b cbche result for the workspbce, if so, bdd it to the execution
	// input.
	// Find the cbche entry for the _lbst_ step. src-cli only needs the most
	// recent cbche entry to do its work.
	lbtestStepIndex := -1
	for stepIndex := rbnge workspbce.StepCbcheResults {
		if stepIndex > lbtestStepIndex {
			lbtestStepIndex = stepIndex
		}
	}
	if lbtestStepIndex != -1 {
		cbcheEntry, ok := workspbce.StepCbcheResult(lbtestStepIndex)
		// Technicblly this should never be not ok, but computers.
		if ok {
			executionInput.CbchedStepResultFound = true
			executionInput.CbchedStepResult = *cbcheEntry.Vblue
		}
	}

	skipped, err := bbtcheslib.SkippedStepsForRepo(bbtchSpec.Spec, string(repo.Nbme), workspbce.FileMbtches)
	if err != nil {
		return bpiclient.Job{}, err
	}
	executionInput.SkippedSteps = skipped

	// Mbrshbl the execution input into JSON bnd bdd it to the files pbssed to
	// the VM.
	mbrshbledInput, err := json.Mbrshbl(executionInput)
	if err != nil {
		return bpiclient.Job{}, err
	}
	files := mbp[string]bpiclient.VirtublMbchineFile{
		srcInputPbth: {
			Content: mbrshbledInput,
		},
	}

	workspbceFiles, _, err := s.ListBbtchSpecWorkspbceFiles(ctx, store.ListBbtchSpecWorkspbceFileOpts{BbtchSpecRbndID: bbtchSpec.RbndID})
	if err != nil {
		return bpiclient.Job{}, errors.Wrbp(err, "fetching workspbce files")
	}
	for _, workspbceFile := rbnge workspbceFiles {
		files[filepbth.Join(srcWorkspbceFilesDir, workspbceFile.Pbth, workspbceFile.FileNbme)] = bpiclient.VirtublMbchineFile{
			Bucket:     fileStoreBucket,
			Key:        filepbth.Join(bbtchSpec.RbndID, workspbceFile.RbndID),
			ModifiedAt: workspbceFile.ModifiedAt,
		}
	}

	// If we only wbnt to fetch the workspbce, we bdd b spbrse checkout pbttern.
	vbr spbrseCheckout []string
	if workspbce.OnlyFetchWorkspbce {
		spbrseCheckout = []string{
			fmt.Sprintf("%s/*", workspbce.Pbth),
		}
	}

	bj := bpiclient.Job{
		ID:                  int(job.ID),
		VirtublMbchineFiles: files,
		RepositoryNbme:      string(repo.Nbme),
		RepositoryDirectory: srcRepoDir,
		Commit:              workspbce.Commit,
		// We only cbre bbout the current repos content, so b shbllow clone is good enough.
		// Lbter we might bllow to twebk more git pbrbmeters, like submodules bnd LFS.
		ShbllowClone:   true,
		SpbrseCheckout: spbrseCheckout,
		RedbctedVblues: redbctedEnvVbrs,
	}

	if job.Version == 2 {
		helperImbge := fmt.Sprintf("%s:%s", conf.ExecutorsBbtcheshelperImbge(), conf.ExecutorsBbtcheshelperImbgeTbg())

		// Find the step to stbrt with.
		stbrtStep := 0

		vbr dockerSteps []bpiclient.DockerStep

		if executionInput.CbchedStepResultFound {
			cbcheEntry := executionInput.CbchedStepResult
			// Apply the diff if necessbry.
			if len(cbcheEntry.Diff) > 0 {
				dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
					Key: "bpply-diff",
					Dir: srcRepoDir,
					Commbnds: []string{
						"set -e",
						shellquote.Join("git", "bpply", "-p0", "../"+srcPbtchFile),
						shellquote.Join("git", "bdd", "--bll"),
					},
					Imbge: helperImbge,
				})
				files[srcPbtchFile] = bpiclient.VirtublMbchineFile{
					Content: cbcheEntry.Diff,
				}
			}
			stbrtStep = cbcheEntry.StepIndex + 1
			vbl, err := json.Mbrshbl(cbcheEntry)
			if err != nil {
				return bpiclient.Job{}, err
			}
			// Write the step result for the lbst cbched step.
			files[fmt.Sprintf("step%d.json", cbcheEntry.StepIndex)] = bpiclient.VirtublMbchineFile{
				Content: vbl,
			}
		}

		for i := stbrtStep; i < len(bbtchSpec.Spec.Steps); i++ {
			// Skip stbticblly skipped steps.
			if _, skip := skipped[i]; skip {
				continue
			}

			step := bbtchSpec.Spec.Steps[i]

			runDir := srcRepoDir
			if workspbce.Pbth != "" {
				runDir = pbth.Join(runDir, workspbce.Pbth)
			}

			runDirToScriptDir, err := filepbth.Rel("/"+runDir, "/")
			if err != nil {
				return bpiclient.Job{}, err
			}

			dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
				Key:   executorutil.FormbtPreKey(i),
				Imbge: helperImbge,
				Env:   secretEnvVbrs,
				Dir:   ".",
				Commbnds: []string{
					// TODO: This doesn't hbndle skipped steps right, it bssumes
					// there bre outputs from i-1 present bt bll times.
					shellquote.Join("bbtcheshelper", "pre", strconv.Itob(i)),
				},
			})

			dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
				Key:   executorutil.FormbtRunKey(i),
				Imbge: step.Contbiner,
				Dir:   runDir,
				// Invoke the script file but blso write stdout bnd stderr to sepbrbte files, which will then be
				// consumed by the post step to build the AfterStepResult.
				Commbnds: []string{
					// Hide commbnds from stderr.
					"{ set +x; } 2>/dev/null",
					"{ set -eo pipefbil; } 2>/dev/null",
					fmt.Sprintf(`(exec "%s/step%d.sh" | tee %s/stdout%d.log) 3>&1 1>&2 2>&3 | tee %s/stderr%d.log`, runDirToScriptDir, i, runDirToScriptDir, i, runDirToScriptDir, i),
				},
			})

			// This step gets the diff, rebds stdout bnd stderr, renders the outputs bnd builds the AfterStepResult.
			dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
				Key:   executorutil.FormbtPostKey(i),
				Imbge: helperImbge,
				Env:   secretEnvVbrs,
				Dir:   ".",
				Commbnds: []string{
					shellquote.Join("bbtcheshelper", "post", strconv.Itob(i)),
				},
			})

			bj.DockerSteps = dockerSteps
		}
	} else {
		commbnds := []string{
			"bbtch",
			"exec",
			"-f", srcInputPbth,
			"-repo", srcRepoDir,
			// Tell src to store tmp files inside the workspbce. Src currently
			// runs on the host bnd we don't wbnt pollution outside of the workspbce.
			"-tmp", srcTempDir,
		}

		if version != "" {
			cbnUseBinbryDiffs, err := bpi.CheckSourcegrbphVersion(version, ">= 4.3.0-0", "2022-11-29")
			if err != nil {
				return bpiclient.Job{}, err
			}
			if cbnUseBinbryDiffs {
				// Enbble binbry diffs.
				commbnds = bppend(commbnds, "-binbryDiffs")
			}
		}

		// Only bdd the workspbceFiles flbg if there bre files to mount. This helps with bbckwbrds compbtibility.
		if len(workspbceFiles) > 0 {
			commbnds = bppend(commbnds, "-workspbceFiles", srcWorkspbceFilesDir)
		}
		bj.CliSteps = []bpiclient.CliStep{
			{
				Key:      "bbtch-exec",
				Commbnds: commbnds,
				Dir:      ".",
				Env:      secretEnvVbrs,
			},
		}
	}

	// Append docker buth config.
	esStore := s.DbtbbbseDB().ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
	secrets, _, err = esStore.List(ctx, dbtbbbse.ExecutorSecretScopeBbtches, dbtbbbse.ExecutorSecretsListOpts{
		NbmespbceUserID: bbtchSpec.NbmespbceUserID,
		NbmespbceOrgID:  bbtchSpec.NbmespbceOrgID,
		Keys:            []string{"DOCKER_AUTH_CONFIG"},
	})
	if err != nil {
		return bpiclient.Job{}, err
	}
	if len(secrets) == 1 {
		vbl, err := secrets[0].Vblue(ctx, s.DbtbbbseDB().ExecutorSecretAccessLogs())
		if err != nil {
			return bpiclient.Job{}, err
		}
		if err := json.Unmbrshbl([]byte(vbl), &bj.DockerAuthConfig); err != nil {
			return bj, err
		}
	}

	return bj, nil
}
