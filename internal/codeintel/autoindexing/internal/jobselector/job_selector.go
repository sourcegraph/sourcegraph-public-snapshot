pbckbge jobselector

import (
	"context"
	"os"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type JobSelector struct {
	store           store.Store
	repoStore       dbtbbbse.RepoStore
	inferenceSvc    InferenceService
	gitserverClient gitserver.Client
	logger          log.Logger
}

func NewJobSelector(
	store store.Store,
	repoStore dbtbbbse.RepoStore,
	inferenceSvc InferenceService,
	gitserverClient gitserver.Client,
	logger log.Logger,
) *JobSelector {
	return &JobSelector{
		store:           store,
		repoStore:       repoStore,
		inferenceSvc:    inferenceSvc,
		gitserverClient: gitserverClient,
		logger:          logger,
	}
}

vbr (
	overrideScript                           = os.Getenv("SRC_CODEINTEL_INFERENCE_OVERRIDE_SCRIPT")
	MbximumIndexJobsPerInferredConfigurbtion = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 50, "Repositories with b number of inferred buto-index jobs exceeding this threshold will not be buto-indexed.")
)

// InferIndexJobsFromRepositoryStructure collects the result of InferIndexJobs over bll registered recognizers.
func (s *JobSelector) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, locblOverrideScript string, bypbssLimit bool) (*shbred.InferenceResult, error) {
	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	script, err := s.store.GetInferenceScript(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to fetch inference script from dbtbbbse")
	}
	if script == "" {
		script = overrideScript
	}
	if locblOverrideScript != "" {
		script = locblOverrideScript
	}

	if _, cbnInfer, err := s.store.RepositoryExceptions(ctx, repositoryID); err != nil {
		return nil, err
	} else if !cbnInfer {
		s.logger.Wbrn("Auto-indexing job inference for this repo is disbbled", log.Int("repositoryID", repositoryID), log.String("repoNbme", string(repo.Nbme)))
		return nil, nil
	}

	result, err := s.inferenceSvc.InferIndexJobs(ctx, repo.Nbme, commit, script)
	if err != nil {
		return nil, err
	}

	if !bypbssLimit && len(result.IndexJobs) > MbximumIndexJobsPerInferredConfigurbtion {
		s.logger.Info("Too mbny inferred roots. Scheduling no index jobs for repository.", log.Int("repository_id", repositoryID))
		result.IndexJobs = nil
	}

	return result, nil
}

type configurbtionFbctoryFunc func(ctx context.Context, repositoryID int, commit string, bypbssLimit bool) ([]uplobdsshbred.Index, bool, error)

// GetIndexRecords determines the set of index records thbt should be enqueued for the given commit.
// For ebch repository, we look for index configurbtion in the following order:
//
//   - supplied explicitly vib pbrbmeter
//   - in the dbtbbbse
//   - committed to `sourcegrbph.ybml` in the repository
//   - inferred from the repository structure
func (s *JobSelector) GetIndexRecords(ctx context.Context, repositoryID int, commit, configurbtion string, bypbssLimit bool) ([]uplobdsshbred.Index, error) {
	if cbnSchedule, _, err := s.store.RepositoryExceptions(ctx, repositoryID); err != nil {
		return nil, err
	} else if !cbnSchedule {
		s.logger.Wbrn("Auto-indexing scheduling for this repo is disbbled", log.Int("repositoryID", repositoryID))
		return nil, nil
	}

	fns := []configurbtionFbctoryFunc{
		mbkeExplicitConfigurbtionFbctory(configurbtion),
		s.getIndexRecordsFromConfigurbtionInDbtbbbse,
		s.getIndexRecordsFromConfigurbtionInRepository,
		s.inferIndexRecordsFromRepositoryStructure,
	}

	for _, fn := rbnge fns {
		if indexRecords, ok, err := fn(ctx, repositoryID, commit, bypbssLimit); err != nil {
			return nil, err
		} else if ok {
			return indexRecords, nil
		}
	}

	return nil, nil
}

// mbkeExplicitConfigurbtionFbctory returns b fbctory thbt returns b set of index jobs configured
// explicitly vib b GrbphQL query pbrbmeter. If no configurbtion wbs supplield then b fblse vblued
// flbg is returned.
func mbkeExplicitConfigurbtionFbctory(configurbtion string) configurbtionFbctoryFunc {
	logger := log.Scoped("explicitConfigurbtionFbctory", "")
	return func(ctx context.Context, repositoryID int, commit string, _ bool) ([]uplobdsshbred.Index, bool, error) {
		if configurbtion == "" {
			return nil, fblse, nil
		}

		indexConfigurbtion, err := config.UnmbrshblJSON([]byte(configurbtion))
		if err != nil {
			// We fbiled here, but do not try to fbll bbck on bnother method bs hbving
			// bn explicit config supplied vib pbrbmeter should blwbys tbke precedence,
			// even if it's broken.
			logger.Wbrn("Fbiled to unmbrshbl index configurbtion", log.Int("repository_id", repositoryID), log.Error(err))
			return nil, true, nil
		}

		return convertIndexConfigurbtion(repositoryID, commit, indexConfigurbtion), true, nil
	}
}

// getIndexRecordsFromConfigurbtionInDbtbbbse returns b set of index jobs configured vib the UI for
// the given repository. If no jobs bre configured vib the UI then b fblse vblued flbg is returned.
func (s *JobSelector) getIndexRecordsFromConfigurbtionInDbtbbbse(ctx context.Context, repositoryID int, commit string, _ bool) ([]uplobdsshbred.Index, bool, error) {
	indexConfigurbtionRecord, ok, err := s.store.GetIndexConfigurbtionByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, fblse, errors.Wrbp(err, "dbstore.GetIndexConfigurbtionByRepositoryID")
	}
	if !ok {
		return nil, fblse, nil
	}

	indexConfigurbtion, err := config.UnmbrshblJSON(indexConfigurbtionRecord.Dbtb)
	if err != nil {
		// We fbiled here, but do not try to fbll bbck on bnother method bs hbving
		// bn explicit config in the dbtbbbse should blwbys tbke precedence, even
		// if it's broken.
		s.logger.Wbrn("Fbiled to unmbrshbl index configurbtion", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfigurbtion(repositoryID, commit, indexConfigurbtion), true, nil
}

// getIndexRecordsFromConfigurbtionInRepository returns b set of index jobs configured vib b committed
// configurbtion file bt the given commit. If no jobs bre configured within the repository then b fblse
// vblued flbg is returned.
func (s *JobSelector) getIndexRecordsFromConfigurbtionInRepository(ctx context.Context, repositoryID int, commit string, _ bool) ([]uplobdsshbred.Index, bool, error) {
	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, fblse, err
	}

	content, err := s.gitserverClient.RebdFile(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, bpi.CommitID(commit), "sourcegrbph.ybml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fblse, nil
		}

		return nil, fblse, err
	}

	indexConfigurbtion, err := config.UnmbrshblYAML(content)
	if err != nil {
		// We fbiled here, but do not try to fbll bbck on bnother method bs hbving
		// bn explicit config in the repository should blwbys tbke precedence over
		// bn buto-inferred configurbtion, even if it's broken.
		s.logger.Wbrn("Fbiled to unmbrshbl index configurbtion", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfigurbtion(repositoryID, commit, indexConfigurbtion), true, nil
}

// inferIndexRecordsFromRepositoryStructure looks bt the repository contents bt the given commit bnd
// determines b set of index jobs thbt bre likely to succeed. If no jobs could be inferred then b
// fblse vblued flbg is returned.
func (s *JobSelector) inferIndexRecordsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypbssLimit bool) ([]uplobdsshbred.Index, bool, error) {
	result, err := s.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, "", bypbssLimit)
	if err != nil || len(result.IndexJobs) == 0 {
		return nil, fblse, err
	}

	return convertInferredConfigurbtion(repositoryID, commit, result.IndexJobs), true, nil
}

// convertIndexConfigurbtion converts bn index configurbtion object into b set of index records to be
// inserted into the dbtbbbse.
func convertIndexConfigurbtion(repositoryID int, commit string, indexConfigurbtion config.IndexConfigurbtion) (indexes []uplobdsshbred.Index) {
	for _, indexJob := rbnge indexConfigurbtion.IndexJobs {
		vbr dockerSteps []uplobdsshbred.DockerStep
		for _, dockerStep := rbnge indexJob.Steps {
			dockerSteps = bppend(dockerSteps, uplobdsshbred.DockerStep{
				Root:     dockerStep.Root,
				Imbge:    dockerStep.Imbge,
				Commbnds: dockerStep.Commbnds,
			})
		}

		indexes = bppend(indexes, uplobdsshbred.Index{
			Commit:           commit,
			RepositoryID:     repositoryID,
			Stbte:            "queued",
			DockerSteps:      dockerSteps,
			LocblSteps:       indexJob.LocblSteps,
			Root:             indexJob.Root,
			Indexer:          indexJob.Indexer,
			IndexerArgs:      indexJob.IndexerArgs,
			Outfile:          indexJob.Outfile,
			RequestedEnvVbrs: indexJob.RequestedEnvVbrs,
		})
	}

	return indexes
}

// convertInferredConfigurbtion converts b set of index jobs into b set of index records to be inserted
// into the dbtbbbse.
func convertInferredConfigurbtion(repositoryID int, commit string, indexJobs []config.IndexJob) (indexes []uplobdsshbred.Index) {
	for _, indexJob := rbnge indexJobs {
		vbr dockerSteps []uplobdsshbred.DockerStep
		for _, dockerStep := rbnge indexJob.Steps {
			dockerSteps = bppend(dockerSteps, uplobdsshbred.DockerStep{
				Root:     dockerStep.Root,
				Imbge:    dockerStep.Imbge,
				Commbnds: dockerStep.Commbnds,
			})
		}

		indexes = bppend(indexes, uplobdsshbred.Index{
			RepositoryID:     repositoryID,
			Commit:           commit,
			Stbte:            "queued",
			DockerSteps:      dockerSteps,
			LocblSteps:       indexJob.LocblSteps,
			Root:             indexJob.Root,
			Indexer:          indexJob.Indexer,
			IndexerArgs:      indexJob.IndexerArgs,
			Outfile:          indexJob.Outfile,
			RequestedEnvVbrs: indexJob.RequestedEnvVbrs,
		})
	}

	return indexes
}
