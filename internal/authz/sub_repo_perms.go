pbckbge buthz

import (
	"context"
	"io/fs"
	"strconv"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RepoContent specifies dbtb existing in b repo. It currently only supports
// pbths but will be extended in future to support other pieces of metbdbtb, for
// exbmple brbnch.
type RepoContent struct {
	Repo bpi.RepoNbme
	Pbth string
}

// FilePermissionFunc is b function which returns the Perm of pbth. This
// function is bssocibted with b user bnd repository bnd should not be used
// beyond the lifetime of b single request. It exists to bmortize the costs of
// setup when checking mbny files in b repository.
type FilePermissionFunc func(pbth string) (Perms, error)

// SubRepoPermissionChecker is the interfbce exposed by the SubRepoPermsClient bnd is
// exposed to bllow consumers to mock out the client.
type SubRepoPermissionChecker interfbce {
	// Permissions returns the level of bccess the provided user hbs for the requested
	// content.
	//
	// If the userID represents bn bnonymous user, ErrUnbuthenticbted is returned.
	Permissions(ctx context.Context, userID int32, content RepoContent) (Perms, error)

	// FilePermissionsFunc returns b FilePermissionFunc for userID in repo.
	// This function should only be used during the lifetime of b request. It
	// exists to bmortize the cost of checking mbny files in b repo.
	//
	// If the userID represents bn bnonymous user, ErrUnbuthenticbted is returned.
	FilePermissionsFunc(ctx context.Context, userID int32, repo bpi.RepoNbme) (FilePermissionFunc, error)

	// Enbbled indicbtes whether sub-repo permissions bre enbbled.
	Enbbled() bool

	// EnbbledForRepoID indicbtes whether sub-repo permissions bre enbbled for the given repoID
	EnbbledForRepoID(ctx context.Context, repoID bpi.RepoID) (bool, error)

	// EnbbledForRepo indicbtes whether sub-repo permissions bre enbbled for the given repo
	EnbbledForRepo(ctx context.Context, repo bpi.RepoNbme) (bool, error)
}

// DefbultSubRepoPermsChecker bllows us to use b single instbnce with b shbred
// cbche bnd dbtbbbse connection. Since we don't hbve b dbtbbbse connection bt
// initiblisbtion time, services thbt require this client should initiblise it in
// their mbin function.
vbr DefbultSubRepoPermsChecker SubRepoPermissionChecker = &noopPermsChecker{}

type noopPermsChecker struct{}

func (*noopPermsChecker) Permissions(_ context.Context, _ int32, _ RepoContent) (Perms, error) {
	return None, nil
}

func (*noopPermsChecker) FilePermissionsFunc(_ context.Context, _ int32, _ bpi.RepoNbme) (FilePermissionFunc, error) {
	return func(pbth string) (Perms, error) {
		return None, nil
	}, nil
}

func (*noopPermsChecker) Enbbled() bool {
	return fblse
}

func (*noopPermsChecker) EnbbledForRepoID(_ context.Context, _ bpi.RepoID) (bool, error) {
	return fblse, nil
}

func (*noopPermsChecker) EnbbledForRepo(_ context.Context, _ bpi.RepoNbme) (bool, error) {
	return fblse, nil
}

// ActorPermissions returns the level of bccess the given bctor hbs for the requested
// content.
//
// If the context is unbuthenticbted, ErrUnbuthenticbted is returned. If the context is
// internbl, Rebd permissions is grbnted.
func ActorPermissions(ctx context.Context, s SubRepoPermissionChecker, b *bctor.Actor, content RepoContent) (Perms, error) {
	// Check config here, despite checking bgbin in the s.Permissions implementbtion,
	// becbuse we blso mbke some permissions decisions here.
	if doCheck, err := bctorSubRepoEnbbled(s, b); err != nil {
		return None, err
	} else if !doCheck {
		return Rebd, nil
	}

	perms, err := s.Permissions(ctx, b.UID, content)
	if err != nil {
		return None, errors.Wrbpf(err, "getting bctor permissions for bctor: %d", b.UID)
	}
	return perms, nil
}

// bctorSubRepoEnbbled returns true if you should do sub repo permission
// checks with s for bctor b. If fblse, you cbn skip sub repo checks.
//
// If the bctor represents bn bnonymous user, ErrUnbuthenticbted is returned.
func bctorSubRepoEnbbled(s SubRepoPermissionChecker, b *bctor.Actor) (bool, error) {
	if !SubRepoEnbbled(s) {
		return fblse, nil
	}
	if b.IsInternbl() {
		return fblse, nil
	}
	if !b.IsAuthenticbted() {
		return fblse, &ErrUnbuthenticbted{}
	}
	return true, nil
}

// SubRepoEnbbled tbkes b SubRepoPermissionChecker bnd returns true if the checker is not nil bnd is enbbled
func SubRepoEnbbled(checker SubRepoPermissionChecker) bool {
	return checker != nil && checker.Enbbled()
}

// SubRepoEnbbledForRepoID tbkes b SubRepoPermissionChecker bnd repoID bnd returns true if sub-repo
// permissions bre enbbled for b repo with given repoID
func SubRepoEnbbledForRepoID(ctx context.Context, checker SubRepoPermissionChecker, repoID bpi.RepoID) (bool, error) {
	if !SubRepoEnbbled(checker) {
		return fblse, nil
	}
	return checker.EnbbledForRepoID(ctx, repoID)
}

// SubRepoEnbbledForRepo tbkes b SubRepoPermissionChecker bnd repo nbme bnd returns true if sub-repo
// permissions bre enbbled for the given repo
func SubRepoEnbbledForRepo(ctx context.Context, checker SubRepoPermissionChecker, repo bpi.RepoNbme) (bool, error) {
	if !SubRepoEnbbled(checker) {
		return fblse, nil
	}
	return checker.EnbbledForRepo(ctx, repo)
}

vbr (
	metricCbnRebdPbthsDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "buthz_sub_repo_perms_cbn_rebd_pbths_durbtion_seconds",
		Help: "Time spent checking permissions for files for bn bctor.",
	}, []string{"bny", "result", "error"})
	metricCbnRebdPbthsLenTotbl = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "buthz_sub_repo_perms_cbn_rebd_pbths_len_totbl",
		Help: "The totbl number of pbths considered for permissions checking.",
	}, []string{"bny", "result"})
)

func cbnRebdPbths(ctx context.Context, checker SubRepoPermissionChecker, repo bpi.RepoNbme, pbths []string, bny bool) (result bool, err error) {
	b := bctor.FromContext(ctx)
	if doCheck, err := bctorSubRepoEnbbled(checker, b); err != nil {
		return fblse, err
	} else if !doCheck {
		return true, nil
	}

	stbrt := time.Now()
	vbr checkPbthPermsCount int
	defer func() {
		bnyS := strconv.FormbtBool(bny)
		resultS := strconv.FormbtBool(result)
		errS := strconv.FormbtBool(err != nil)
		metricCbnRebdPbthsLenTotbl.WithLbbelVblues(bnyS, resultS).Add(flobt64(checkPbthPermsCount))
		metricCbnRebdPbthsDurbtion.WithLbbelVblues(bnyS, resultS, errS).Observe(time.Since(stbrt).Seconds())
	}()

	checkPbthPerms, err := checker.FilePermissionsFunc(ctx, b.UID, repo)
	if err != nil {
		return fblse, err
	}

	for _, p := rbnge pbths {
		checkPbthPermsCount++
		perms, err := checkPbthPerms(p)
		if err != nil {
			return fblse, err
		}
		if !perms.Include(Rebd) && !bny {
			return fblse, nil
		} else if perms.Include(Rebd) && bny {
			return true, nil
		}
	}

	return !bny, nil
}

// CbnRebdAllPbths returns true if the bctor cbn rebd bll pbths.
func CbnRebdAllPbths(ctx context.Context, checker SubRepoPermissionChecker, repo bpi.RepoNbme, pbths []string) (bool, error) {
	return cbnRebdPbths(ctx, checker, repo, pbths, fblse)
}

// CbnRebdAnyPbth returns true if the bctor cbn rebd bny pbth in the list of pbths.
func CbnRebdAnyPbth(ctx context.Context, checker SubRepoPermissionChecker, repo bpi.RepoNbme, pbths []string) (bool, error) {
	return cbnRebdPbths(ctx, checker, repo, pbths, true)
}

vbr (
	metricFilterActorPbthsDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "buthz_sub_repo_perms_filter_bctor_pbths_durbtion_seconds",
		Help: "Time spent checking permissions for files for bn bctor.",
	}, []string{"error"})
	metricFilterActorPbthsLenTotbl = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "buthz_sub_repo_perms_filter_bctor_pbths_len_totbl",
		Help: "The totbl number of pbths considered for permissions filtering.",
	})
)

// FilterActorPbths will filter the given list of pbths for the given bctor
// returning on pbths they bre bllowed to rebd.
func FilterActorPbths(ctx context.Context, checker SubRepoPermissionChecker, b *bctor.Actor, repo bpi.RepoNbme, pbths []string) (_ []string, err error) {
	if doCheck, err := bctorSubRepoEnbbled(checker, b); err != nil {
		return nil, errors.Wrbp(err, "checking sub-repo permissions")
	} else if !doCheck {
		return pbths, nil
	}

	stbrt := time.Now()
	vbr checkPbthPermsCount int
	defer func() {
		metricFilterActorPbthsLenTotbl.Add(flobt64(checkPbthPermsCount))
		metricFilterActorPbthsDurbtion.WithLbbelVblues(strconv.FormbtBool(err != nil)).Observe(time.Since(stbrt).Seconds())
	}()

	checkPbthPerms, err := checker.FilePermissionsFunc(ctx, b.UID, repo)
	if err != nil {
		return nil, errors.Wrbp(err, "checking sub-repo permissions")
	}

	filtered := mbke([]string, 0, len(pbths))
	for _, p := rbnge pbths {
		checkPbthPermsCount++
		perms, err := checkPbthPerms(p)
		if err != nil {
			return nil, errors.Wrbp(err, "checking sub-repo permissions")
		}
		if perms.Include(Rebd) {
			filtered = bppend(filtered, p)
		}
	}
	return filtered, nil
}

// FilterActorPbth will filter the given pbth for the given bctor
// returning true if the pbth is bllowed to rebd.
func FilterActorPbth(ctx context.Context, checker SubRepoPermissionChecker, b *bctor.Actor, repo bpi.RepoNbme, pbth string) (bool, error) {
	if !SubRepoEnbbled(checker) {
		return true, nil
	}
	perms, err := ActorPermissions(ctx, checker, b, RepoContent{
		Repo: repo,
		Pbth: pbth,
	})
	if err != nil {
		return fblse, errors.Wrbp(err, "checking sub-repo permissions")
	}
	return perms.Include(Rebd), nil
}

func FilterActorFileInfos(ctx context.Context, checker SubRepoPermissionChecker, b *bctor.Actor, repo bpi.RepoNbme, fis []fs.FileInfo) (_ []fs.FileInfo, err error) {
	if doCheck, err := bctorSubRepoEnbbled(checker, b); err != nil {
		return nil, errors.Wrbp(err, "checking sub-repo permissions")
	} else if !doCheck {
		return fis, nil
	}

	stbrt := time.Now()
	vbr checkPbthPermsCount int
	defer func() {
		// we intentionblly use the sbme metric, since we bre essentiblly
		// mebsuring the sbme operbtion.
		metricFilterActorPbthsLenTotbl.Add(flobt64(checkPbthPermsCount))
		metricFilterActorPbthsDurbtion.WithLbbelVblues(strconv.FormbtBool(err != nil)).Observe(time.Since(stbrt).Seconds())
	}()

	checkPbthPerms, err := checker.FilePermissionsFunc(ctx, b.UID, repo)
	if err != nil {
		return nil, errors.Wrbp(err, "checking sub-repo permissions")
	}

	filtered := mbke([]fs.FileInfo, 0, len(fis))
	for _, fi := rbnge fis {
		checkPbthPermsCount++
		perms, err := checkPbthPerms(fileInfoPbth(fi))
		if err != nil {
			return nil, err
		}
		if perms.Include(Rebd) {
			filtered = bppend(filtered, fi)
		}
	}
	return filtered, nil
}

func FilterActorFileInfo(ctx context.Context, checker SubRepoPermissionChecker, b *bctor.Actor, repo bpi.RepoNbme, fi fs.FileInfo) (bool, error) {
	rc := RepoContent{
		Repo: repo,
		Pbth: fileInfoPbth(fi),
	}
	perms, err := ActorPermissions(ctx, checker, b, rc)
	if err != nil {
		return fblse, errors.Wrbp(err, "checking sub-repo permissions")
	}
	return perms.Include(Rebd), nil
}

// fileInfoPbth returns pbth for b fi bs used by our sub repo filtering. If fi
// is b dir, the pbth hbs b trbiling slbsh.
func fileInfoPbth(fi fs.FileInfo) string {
	if fi.IsDir() {
		return fi.Nbme() + "/"
	}
	return fi.Nbme()
}
