package batches

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"

	"github.com/kballard/go-shellquote"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor/types"
	executorutil "github.com/sourcegraph/sourcegraph/internal/executor/util"
	"github.com/sourcegraph/sourcegraph/lib/api"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	srcInputPath         = "input.json"
	srcPatchFile         = "state.diff"
	srcRepoDir           = "repository"
	srcTempDir           = ".src-tmp"
	srcWorkspaceFilesDir = "workspace-files"
)

type BatchesStore interface {
	GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)
	ListBatchSpecWorkspaceFiles(ctx context.Context, opts store.ListBatchSpecWorkspaceFileOpts) ([]*btypes.BatchSpecWorkspaceFile, int64, error)

	DatabaseDB() database.DB
}

const fileStoreBucket = "batch-changes"

// transformRecord transforms a *btypes.BatchSpecWorkspaceExecutionJob into an apiclient.Job.
func transformRecord(ctx context.Context, logger log.Logger, s BatchesStore, job *btypes.BatchSpecWorkspaceExecutionJob, version string) (apiclient.Job, error) {
	workspace, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return apiclient.Job{}, errors.Wrapf(err, "fetching workspace %d", job.BatchSpecWorkspaceID)
	}

	batchSpec, err := s.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching batch spec")
	}

	// This should never happen. To get some easier debugging when a user sees strange
	// behavior, we log some additional context.
	if job.UserID != batchSpec.UserID {
		logger.Error("bad DB state: batch spec workspace execution job did not have the same user ID as the associated batch spec")
	}

	// 🚨 SECURITY: Set the actor on the context so we check for permissions
	// when loading the repository and getting secret values.
	ctx = actor.WithActor(ctx, actor.FromUser(job.UserID))

	// Next, we fetch all secrets that are requested for the execution.
	rk := batchSpec.Spec.RequiredEnvVars()
	var secrets []*database.ExecutorSecret
	if len(rk) > 0 {
		esStore := s.DatabaseDB().ExecutorSecrets(keyring.Default().ExecutorSecretKey)
		secrets, _, err = esStore.List(ctx, database.ExecutorSecretScopeBatches, database.ExecutorSecretsListOpts{
			NamespaceUserID: batchSpec.NamespaceUserID,
			NamespaceOrgID:  batchSpec.NamespaceOrgID,
			Keys:            rk,
		})
		if err != nil {
			return apiclient.Job{}, err
		}
	}

	// And build the env vars from the secrets.
	secretEnvVars := make([]string, len(secrets))
	redactedEnvVars := make(map[string]string, len(secrets))
	esalStore := s.DatabaseDB().ExecutorSecretAccessLogs()
	for i, secret := range secrets {
		// Get the secret value. This also creates an access log entry in the
		// name of the user.
		val, err := secret.Value(ctx, esalStore)
		if err != nil {
			return apiclient.Job{}, err
		}

		secretEnvVars[i] = fmt.Sprintf("%s=%s", secret.Key, val)
		// We redact secret values as ${{ secrets.NAME }}.
		redactedEnvVars[val] = fmt.Sprintf("${{ secrets.%s }}", secret.Key)
	}

	repo, err := s.DatabaseDB().Repos().Get(ctx, workspace.RepoID)
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching repo")
	}

	executionInput := batcheslib.WorkspacesExecutionInput{
		Repository: batcheslib.WorkspaceRepo{
			ID:   string(graphqlbackend.MarshalRepositoryID(repo.ID)),
			Name: string(repo.Name),
		},
		Branch: batcheslib.WorkspaceBranch{
			Name:   workspace.Branch,
			Target: batcheslib.Commit{OID: workspace.Commit},
		},
		Path:               workspace.Path,
		OnlyFetchWorkspace: workspace.OnlyFetchWorkspace,
		Steps:              batchSpec.Spec.Steps,
		SearchResultPaths:  workspace.FileMatches,
		BatchChangeAttributes: template.BatchChangeAttributes{
			Name:        batchSpec.Spec.Name,
			Description: batchSpec.Spec.Description,
		},
	}

	// Check if we have a cache result for the workspace, if so, add it to the execution
	// input.
	// Find the cache entry for the _last_ step. src-cli only needs the most
	// recent cache entry to do its work.
	latestStepIndex := -1
	for stepIndex := range workspace.StepCacheResults {
		if stepIndex > latestStepIndex {
			latestStepIndex = stepIndex
		}
	}
	if latestStepIndex != -1 {
		cacheEntry, ok := workspace.StepCacheResult(latestStepIndex)
		// Technically this should never be not ok, but computers.
		if ok {
			executionInput.CachedStepResultFound = true
			executionInput.CachedStepResult = *cacheEntry.Value
		}
	}

	skipped, err := batcheslib.SkippedStepsForRepo(batchSpec.Spec, string(repo.Name), workspace.FileMatches)
	if err != nil {
		return apiclient.Job{}, err
	}
	executionInput.SkippedSteps = skipped

	// Marshal the execution input into JSON and add it to the files passed to
	// the VM.
	marshaledInput, err := json.Marshal(executionInput)
	if err != nil {
		return apiclient.Job{}, err
	}
	files := map[string]apiclient.VirtualMachineFile{
		srcInputPath: {
			Content: marshaledInput,
		},
	}

	workspaceFiles, _, err := s.ListBatchSpecWorkspaceFiles(ctx, store.ListBatchSpecWorkspaceFileOpts{BatchSpecRandID: batchSpec.RandID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching workspace files")
	}
	for _, workspaceFile := range workspaceFiles {
		files[filepath.Join(srcWorkspaceFilesDir, workspaceFile.Path, workspaceFile.FileName)] = apiclient.VirtualMachineFile{
			Bucket:     fileStoreBucket,
			Key:        filepath.Join(batchSpec.RandID, workspaceFile.RandID),
			ModifiedAt: workspaceFile.ModifiedAt,
		}
	}

	// If we only want to fetch the workspace, we add a sparse checkout pattern.
	var sparseCheckout []string
	if workspace.OnlyFetchWorkspace {
		sparseCheckout = []string{
			fmt.Sprintf("%s/*", workspace.Path),
		}
	}

	aj := apiclient.Job{
		ID:                  int(job.ID),
		VirtualMachineFiles: files,
		RepositoryName:      string(repo.Name),
		RepositoryDirectory: srcRepoDir,
		Commit:              workspace.Commit,
		// We only care about the current repos content, so a shallow clone is good enough.
		// Later we might allow to tweak more git parameters, like submodules and LFS.
		ShallowClone:   true,
		SparseCheckout: sparseCheckout,
		RedactedValues: redactedEnvVars,
	}

	if job.Version == 2 {
		helperImage := fmt.Sprintf("%s:%s", conf.ExecutorsBatcheshelperImage(), conf.ExecutorsBatcheshelperImageTag())

		// Find the step to start with.
		startStep := 0

		var dockerSteps []apiclient.DockerStep

		if executionInput.CachedStepResultFound {
			cacheEntry := executionInput.CachedStepResult
			// Apply the diff if necessary.
			if len(cacheEntry.Diff) > 0 {
				dockerSteps = append(dockerSteps, apiclient.DockerStep{
					Key: "apply-diff",
					Dir: srcRepoDir,
					Commands: []string{
						"set -e",
						shellquote.Join("git", "apply", "-p0", "../"+srcPatchFile),
						shellquote.Join("git", "add", "--all"),
					},
					Image: helperImage,
				})
				files[srcPatchFile] = apiclient.VirtualMachineFile{
					Content: cacheEntry.Diff,
				}
			}
			startStep = cacheEntry.StepIndex + 1
			val, err := json.Marshal(cacheEntry)
			if err != nil {
				return apiclient.Job{}, err
			}
			// Write the step result for the last cached step.
			files[fmt.Sprintf("step%d.json", cacheEntry.StepIndex)] = apiclient.VirtualMachineFile{
				Content: val,
			}
		}

		for i := startStep; i < len(batchSpec.Spec.Steps); i++ {
			// Skip statically skipped steps.
			if _, skip := skipped[i]; skip {
				continue
			}

			step := batchSpec.Spec.Steps[i]

			runDir := srcRepoDir
			if workspace.Path != "" {
				runDir = path.Join(runDir, workspace.Path)
			}

			runDirToScriptDir, err := filepath.Rel("/"+runDir, "/")
			if err != nil {
				return apiclient.Job{}, err
			}

			dockerSteps = append(dockerSteps, apiclient.DockerStep{
				Key:   executorutil.FormatPreKey(i),
				Image: helperImage,
				Env:   secretEnvVars,
				Dir:   ".",
				Commands: []string{
					// TODO: This doesn't handle skipped steps right, it assumes
					// there are outputs from i-1 present at all times.
					shellquote.Join("batcheshelper", "pre", strconv.Itoa(i)),
				},
			})

			dockerSteps = append(dockerSteps, apiclient.DockerStep{
				Key:   executorutil.FormatRunKey(i),
				Image: step.Container,
				Dir:   runDir,
				// Invoke the script file but also write stdout and stderr to separate files, which will then be
				// consumed by the post step to build the AfterStepResult.
				Commands: []string{
					// Hide commands from stderr.
					"{ set +x; } 2>/dev/null",
					"{ set -eo pipefail; } 2>/dev/null",
					fmt.Sprintf(`(exec "%s/step%d.sh" | tee %s/stdout%d.log) 3>&1 1>&2 2>&3 | tee %s/stderr%d.log`, runDirToScriptDir, i, runDirToScriptDir, i, runDirToScriptDir, i),
				},
			})

			// This step gets the diff, reads stdout and stderr, renders the outputs and builds the AfterStepResult.
			dockerSteps = append(dockerSteps, apiclient.DockerStep{
				Key:   executorutil.FormatPostKey(i),
				Image: helperImage,
				Env:   secretEnvVars,
				Dir:   ".",
				Commands: []string{
					shellquote.Join("batcheshelper", "post", strconv.Itoa(i)),
				},
			})

			aj.DockerSteps = dockerSteps
		}
	} else {
		commands := []string{
			"batch",
			"exec",
			"-f", srcInputPath,
			"-repo", srcRepoDir,
			// Tell src to store tmp files inside the workspace. Src currently
			// runs on the host and we don't want pollution outside of the workspace.
			"-tmp", srcTempDir,
		}

		if version != "" {
			canUseBinaryDiffs, err := api.CheckSourcegraphVersion(version, ">= 4.3.0-0", "2022-11-29")
			if err != nil {
				return apiclient.Job{}, err
			}
			if canUseBinaryDiffs {
				// Enable binary diffs.
				commands = append(commands, "-binaryDiffs")
			}
		}

		// Only add the workspaceFiles flag if there are files to mount. This helps with backwards compatibility.
		if len(workspaceFiles) > 0 {
			commands = append(commands, "-workspaceFiles", srcWorkspaceFilesDir)
		}
		aj.CliSteps = []apiclient.CliStep{
			{
				Key:      "batch-exec",
				Commands: commands,
				Dir:      ".",
				Env:      secretEnvVars,
			},
		}
	}

	// Append docker auth config.
	esStore := s.DatabaseDB().ExecutorSecrets(keyring.Default().ExecutorSecretKey)
	secrets, _, err = esStore.List(ctx, database.ExecutorSecretScopeBatches, database.ExecutorSecretsListOpts{
		NamespaceUserID: batchSpec.NamespaceUserID,
		NamespaceOrgID:  batchSpec.NamespaceOrgID,
		Keys:            []string{"DOCKER_AUTH_CONFIG"},
	})
	if err != nil {
		return apiclient.Job{}, err
	}
	if len(secrets) == 1 {
		val, err := secrets[0].Value(ctx, s.DatabaseDB().ExecutorSecretAccessLogs())
		if err != nil {
			return apiclient.Job{}, err
		}
		if err := json.Unmarshal([]byte(val), &aj.DockerAuthConfig); err != nil {
			return aj, err
		}
	}

	return aj, nil
}
