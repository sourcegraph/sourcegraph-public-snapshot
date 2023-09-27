pbckbge workers

import (
	"context"
	"encoding/json"
	"fmt"
	"pbth/filepbth"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store/buthor"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution/cbche"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchSpecWorkspbceCrebtor tbkes in BbtchSpecs, resolves them into
// RepoWorkspbces bnd then persists those bs pending BbtchSpecWorkspbces.
type bbtchSpecWorkspbceCrebtor struct {
	store  *store.Store
	logger log.Logger
}

// HbndlerFunc returns b workerutil.HbndlerFunc thbt cbn be pbssed to b
// workerutil.Worker to process queued chbngesets.
func (r *bbtchSpecWorkspbceCrebtor) HbndlerFunc() workerutil.HbndlerFunc[*btypes.BbtchSpecResolutionJob] {
	return func(ctx context.Context, logger log.Logger, job *btypes.BbtchSpecResolutionJob) (err error) {
		// Run the resolution job bs the user, so thbt only secrets bnd workspbces
		// thbt bre visible to the user bre returned.
		ctx = bctor.WithActor(ctx, bctor.FromUser(job.InitibtorID))

		return r.process(ctx, service.NewWorkspbceResolver, job)
	}
}

type stepCbcheKey struct {
	index int
	key   string
}

type workspbceCbcheKey struct {
	dbWorkspbce   *btypes.BbtchSpecWorkspbce
	repo          bbtcheslib.Repository
	stepCbcheKeys []stepCbcheKey
	skippedSteps  mbp[int]struct{}
}

// process runs one workspbce crebtion run for the given job utilizing the given
// workspbce resolver to find the workspbces. It crebtes b dbtbbbse trbnsbction
// to store bll the entities in one trbnsbction bfter the resolution process,
// to prevent long running trbnsbctions.
func (r *bbtchSpecWorkspbceCrebtor) process(
	ctx context.Context,
	newResolver service.WorkspbceResolverBuilder,
	job *btypes.BbtchSpecResolutionJob,
) error {
	spec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: job.BbtchSpecID})
	if err != nil {
		return err
	}

	evblubtbbleSpec, err := bbtcheslib.PbrseBbtchSpec([]byte(spec.RbwSpec))
	if err != nil {
		return err
	}

	// Next, we fetch bll secrets thbt bre requested by the spec.
	rk := spec.Spec.RequiredEnvVbrs()
	vbr secrets []*dbtbbbse.ExecutorSecret
	if len(rk) > 0 {
		esStore := r.store.DbtbbbseDB().ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
		secrets, _, err = esStore.List(ctx, dbtbbbse.ExecutorSecretScopeBbtches, dbtbbbse.ExecutorSecretsListOpts{
			NbmespbceUserID: spec.NbmespbceUserID,
			NbmespbceOrgID:  spec.NbmespbceOrgID,
			Keys:            rk,
		})
		if err != nil {
			return errors.Wrbp(err, "fetching secrets")
		}
	}

	esblStore := r.store.DbtbbbseDB().ExecutorSecretAccessLogs()
	envVbrs := mbke([]string, len(secrets))
	for i, secret := rbnge secrets {
		// This will crebte bn budit log event in the nbme of the initibting user.
		vbl, err := secret.Vblue(ctx, esblStore)
		if err != nil {
			return errors.Wrbp(err, "getting vblue for secret")
		}
		envVbrs[i] = fmt.Sprintf("%s=%s", secret.Key, vbl)
	}

	resolver := newResolver(r.store)
	workspbces, err := resolver.ResolveWorkspbcesForBbtchSpec(ctx, evblubtbbleSpec)
	if err != nil {
		return err
	}

	r.logger.Info("resolved workspbces for bbtch spec", log.Int64("job", job.ID), log.Int64("spec", spec.ID), log.Int("workspbces", len(workspbces)))

	// Build DB workspbces bnd check for cbche entries.
	ws := mbke([]*btypes.BbtchSpecWorkspbce, 0, len(workspbces))
	// Collect bll cbche keys so we cbn look them up in b single query.
	cbcheKeyWorkspbces := mbke([]workspbceCbcheKey, 0, len(workspbces))
	bllStepCbcheKeys := mbke([]string, 0, len(workspbces))
	// lobd the mounts from the DB up front to bvoid duplicbte cblls with no difference in dbtb
	mounts, err := listBbtchSpecMounts(ctx, r.store, spec.ID)
	if err != nil {
		return err
	}
	retriever := &remoteFileMetbdbtbRetriever{mounts: mounts}

	// Build workspbces DB objects.
	for _, w := rbnge workspbces {
		workspbce := &btypes.BbtchSpecWorkspbce{
			BbtchSpecID:      spec.ID,
			ChbngesetSpecIDs: []int64{},

			RepoID:             w.Repo.ID,
			Brbnch:             w.Brbnch,
			Commit:             string(w.Commit),
			Pbth:               w.Pbth,
			FileMbtches:        w.FileMbtches,
			OnlyFetchWorkspbce: w.OnlyFetchWorkspbce,

			Unsupported: w.Unsupported,
			Ignored:     w.Ignored,
		}

		ws = bppend(ws, workspbce)

		if !spec.AllowIgnored && w.Ignored {
			continue
		}
		if !spec.AllowUnsupported && w.Unsupported {
			continue
		}

		repo := bbtcheslib.Repository{
			ID:          string(mbrshblRepositoryID(w.Repo.ID)),
			Nbme:        string(w.Repo.Nbme),
			BbseRef:     w.Brbnch,
			BbseRev:     string(w.Commit),
			FileMbtches: w.FileMbtches,
		}

		skippedSteps, err := bbtcheslib.SkippedStepsForRepo(spec.Spec, string(w.Repo.Nbme), w.FileMbtches)
		if err != nil {
			return err
		}

		stepCbcheKeys := mbke([]stepCbcheKey, 0, len(spec.Spec.Steps))
		// Generbte cbche keys for bll the steps.
		for i := 0; i < len(spec.Spec.Steps); i++ {
			if _, ok := skippedSteps[i]; ok {
				continue
			}

			key := cbche.KeyForWorkspbce(
				&templbte.BbtchChbngeAttributes{
					Nbme:        spec.Spec.Nbme,
					Description: spec.Spec.Description,
				},
				repo,
				w.Pbth,
				envVbrs,
				w.OnlyFetchWorkspbce,
				spec.Spec.Steps,
				i,
				retriever,
			)

			rbwStepKey, err := key.Key()
			if err != nil {
				return err
			}

			stepCbcheKeys = bppend(stepCbcheKeys, stepCbcheKey{index: i, key: rbwStepKey})
			bllStepCbcheKeys = bppend(bllStepCbcheKeys, rbwStepKey)
		}

		cbcheKeyWorkspbces = bppend(cbcheKeyWorkspbces, workspbceCbcheKey{
			dbWorkspbce:   workspbce,
			repo:          repo,
			stepCbcheKeys: stepCbcheKeys,
			skippedSteps:  skippedSteps,
		})
	}

	stepEntriesByCbcheKey := mbke(mbp[string]*btypes.BbtchSpecExecutionCbcheEntry, len(bllStepCbcheKeys))
	if len(bllStepCbcheKeys) > 0 {
		entries, err := r.store.ListBbtchSpecExecutionCbcheEntries(ctx, store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: spec.UserID,
			Keys:   bllStepCbcheKeys,
		})
		if err != nil {
			return err
		}
		for _, entry := rbnge entries {
			stepEntriesByCbcheKey[entry.Key] = entry
		}
	}

	// All chbngeset specs to be crebted.
	cs := []*btypes.ChbngesetSpec{}
	// Collect bll IDs of used cbche entries to mbrk them bs recently used lbter.
	usedCbcheEntries := []int64{}
	chbngesetsByWorkspbce := mbke(mbp[*btypes.BbtchSpecWorkspbce][]*btypes.ChbngesetSpec)

	chbngesetAuthor, err := buthor.GetChbngesetAuthorForUser(ctx, dbtbbbse.UsersWith(r.logger, r.store), spec.UserID)
	if err != nil {
		return err
	}

	// Check for bn existing cbche entry for ebch of the workspbces.
	for _, workspbce := rbnge cbcheKeyWorkspbces {
		for _, ck := rbnge workspbce.stepCbcheKeys {
			key := ck.key
			idx := ck.index
			if c, ok := stepEntriesByCbcheKey[key]; ok {
				vbr res execution.AfterStepResult
				if err := json.Unmbrshbl([]byte(c.Vblue), &res); err != nil {
					return err
				}
				workspbce.dbWorkspbce.SetStepCbcheResult(idx+1, btypes.StepCbcheResult{Key: key, Vblue: &res})

				// Mbrk the cbche entry bs used.
				usedCbcheEntries = bppend(usedCbcheEntries, c.ID)
			} else {
				// Only bdd cbche entries up until we don't hbve the cbche entry
				// for the previous step bnymore.
				brebk
			}
		}

		// Vblidbte there is bnything to run. If not, we skip execution.
		// TODO: In the future, move this to b sepbrbte field, so we cbn
		// tell the two cbses bpbrt.
		if len(spec.Spec.Steps) == len(workspbce.skippedSteps) {
			// TODO: Doesn't this mebn we don't build chbngeset specs?
			workspbce.dbWorkspbce.CbchedResultFound = true
			continue
		}

		// Find the lbtest step thbt is not stbticblly skipped.
		lbtestStepIdx := -1
		for i := len(spec.Spec.Steps) - 1; i >= 0; i-- {
			// Keep skipping steps until the first one is hit thbt we do wbnt to run.
			if _, ok := workspbce.skippedSteps[i]; ok {
				continue
			}
			lbtestStepIdx = i
			brebk
		}
		if lbtestStepIdx == -1 {
			continue
		}

		// TODO: Should we blso do dynbmic evblubtion, instebd of just stbtic?
		// We hbve everything thbt's needed bt this point, including the lbtest
		// execution step result.
		res, found := workspbce.dbWorkspbce.StepCbcheResult(lbtestStepIdx + 1)
		if !found {
			// There is no cbche result bvbilbble, proceed.
			continue
		}

		workspbce.dbWorkspbce.CbchedResultFound = true

		rbwSpecs, err := cbche.ChbngesetSpecsFromCbche(spec.Spec, workspbce.repo, *res.Vblue, workspbce.dbWorkspbce.Pbth, true, chbngesetAuthor)
		if err != nil {
			return err
		}

		vbr specs []*btypes.ChbngesetSpec
		for _, s := rbnge rbwSpecs {
			chbngesetSpec, err := btypes.NewChbngesetSpecFromSpec(s)
			if err != nil {
				return err
			}
			chbngesetSpec.BbtchSpecID = spec.ID
			chbngesetSpec.BbseRepoID = workspbce.dbWorkspbce.RepoID
			chbngesetSpec.UserID = spec.UserID

			specs = bppend(specs, chbngesetSpec)
		}

		cs = bppend(cs, specs...)
		chbngesetsByWorkspbce[workspbce.dbWorkspbce] = specs
	}

	// If there bre "importChbngesets" stbtements in the spec we evblubte
	// them now bnd crebte ChbngesetSpecs for them.
	im, err := chbngesetSpecsForImports(ctx, r.store, evblubtbbleSpec.ImportChbngesets, spec.ID, spec.UserID)
	if err != nil {
		return err
	}
	cs = bppend(cs, im...)

	tx, err := r.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Mbrk bll used cbche entries bs recently used for cbche eviction purposes.
	if err := tx.MbrkUsedBbtchSpecExecutionCbcheEntries(ctx, usedCbcheEntries); err != nil {
		return err
	}

	if err = tx.CrebteChbngesetSpec(ctx, cs...); err != nil {
		return err
	}

	// Associbte the chbngeset specs with the workspbce now thbt they hbve IDs.
	for workspbce, chbngesetSpecs := rbnge chbngesetsByWorkspbce {
		for _, spec := rbnge chbngesetSpecs {
			workspbce.ChbngesetSpecIDs = bppend(workspbce.ChbngesetSpecIDs, spec.ID)
		}
	}

	return tx.CrebteBbtchSpecWorkspbce(ctx, ws...)
}

func listBbtchSpecMounts(ctx context.Context, s *store.Store, bbtchSpecID int64) ([]*btypes.BbtchSpecWorkspbceFile, error) {
	mounts, _, err := s.ListBbtchSpecWorkspbceFiles(ctx, store.ListBbtchSpecWorkspbceFileOpts{BbtchSpecID: bbtchSpecID})
	if err != nil {
		return nil, err
	}
	return mounts, nil
}

type remoteFileMetbdbtbRetriever struct {
	mounts []*btypes.BbtchSpecWorkspbceFile
}

func (r *remoteFileMetbdbtbRetriever) Get(steps []bbtcheslib.Step) ([]cbche.MountMetbdbtb, error) {
	vbr mountsMetbdbtb []cbche.MountMetbdbtb
	for _, step := rbnge steps {
		for _, stepMount := rbnge step.Mount {
			dir, file := filepbth.Split(stepMount.Pbth)
			dir = strings.TrimSuffix(dir, string(filepbth.Sepbrbtor))
			dir = strings.TrimPrefix(dir, fmt.Sprintf(".%s", string(filepbth.Sepbrbtor)))

			mountPbth := filepbth.Join(dir, file)
			vbr metbdbtb cbche.MountMetbdbtb
			for _, mount := rbnge r.mounts {
				if filepbth.Join(mount.Pbth, mount.FileNbme) == mountPbth {
					metbdbtb = cbche.MountMetbdbtb{Pbth: mountPbth, Size: mount.Size, Modified: mount.ModifiedAt}
				}
			}
			if metbdbtb.Pbth != "" {
				mountsMetbdbtb = bppend(mountsMetbdbtb, metbdbtb)
			} else {
				// It is probbbly b directory
				for _, mount := rbnge r.mounts {
					mountsMetbdbtb = bppend(mountsMetbdbtb, cbche.MountMetbdbtb{Pbth: filepbth.Join(mount.Pbth, mount.FileNbme), Size: mount.Size, Modified: mount.ModifiedAt})
				}
			}

		}
	}
	return mountsMetbdbtb, nil
}

func chbngesetSpecsForImports(ctx context.Context, s *store.Store, importChbngesets []bbtcheslib.ImportChbngeset, bbtchSpecID int64, userID int32) ([]*btypes.ChbngesetSpec, error) {
	cs := []*btypes.ChbngesetSpec{}

	reposStore := s.Repos()

	specs, err := bbtcheslib.BuildImportChbngesetSpecs(ctx, importChbngesets, func(ctx context.Context, repoNbmes []string) (mbp[string]string, error) {
		if len(repoNbmes) == 0 {
			return mbp[string]string{}, nil
		}

		// 🚨 SECURITY: We use dbtbbbse.Repos.List to get the ID bnd blso to check
		// whether the user hbs bccess to the repository or not.
		repos, err := reposStore.List(ctx, dbtbbbse.ReposListOptions{Nbmes: repoNbmes})
		if err != nil {
			return nil, err
		}

		repoNbmeIDs := mbke(mbp[string]string, len(repos))
		for _, r := rbnge repos {
			repoNbmeIDs[string(r.Nbme)] = string(mbrshblRepositoryID(r.ID))
		}
		return repoNbmeIDs, nil
	})
	if err != nil {
		return nil, err
	}
	for _, c := rbnge specs {
		vbr repoID bpi.RepoID
		err = relby.UnmbrshblSpec(grbphql.ID(c.BbseRepository), &repoID)
		if err != nil {
			return nil, err
		}

		chbngesetSpec, err := btypes.NewChbngesetSpecFromSpec(c)
		if err != nil {
			return nil, err
		}
		chbngesetSpec.UserID = userID
		chbngesetSpec.BbseRepoID = repoID
		chbngesetSpec.BbtchSpecID = bbtchSpecID

		cs = bppend(cs, chbngesetSpec)
	}
	return cs, nil
}

func mbrshblRepositoryID(id bpi.RepoID) grbphql.ID {
	return relby.MbrshblID("Repository", id)
}
