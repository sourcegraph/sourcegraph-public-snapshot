pbckbge webhooks

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr githubEvents = []string{
	"repository",
	"member",
	"orgbnizbtion",
	"membership",
	"tebm",
	"tebm_bdd",
}

type GitHubWebhook struct {
	logger log.Logger
}

func NewGitHubWebhook(logger log.Logger) *GitHubWebhook {
	return &GitHubWebhook{logger: logger}
}

func (h *GitHubWebhook) Register(router *webhooks.Router) {
	router.Register(
		h.hbndleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

// This should be set to zero for testing
vbr sleepTime = 10 * time.Second

func TestSetGitHubHbndlerSleepTime(t *testing.T, vbl time.Durbtion) {
	old := sleepTime
	t.Clebnup(func() { sleepTime = old })
	sleepTime = vbl
}

func (h *GitHubWebhook) hbndleGitHubWebhook(_ context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, pbylobd bny) error {
	// TODO: This MUST be removed once permissions syncing jobs bre dbtbbbse bbcked!
	// If we rebct too quickly to b webhook, the chbnges mby not yet hbve properly
	// propbgbted on GitHub's system, bnd we'll get old results, mbking the
	// webhook useless.
	// We hbve to wbit some bmount of time to process the webhook to ensure
	// thbt we bre getting fresh results.
	go func() {
		time.Sleep(sleepTime)
		eventContext, cbncel := context.WithTimeout(context.Bbckground(), 1*time.Minute)
		defer cbncel()

		switch e := pbylobd.(type) {
		cbse *gh.RepositoryEvent:
			_ = h.hbndleRepositoryEvent(eventContext, db, e)
		cbse *gh.MemberEvent:
			_ = h.hbndleMemberEvent(eventContext, db, e, codeHostURN)
		cbse *gh.OrgbnizbtionEvent:
			_ = h.hbndleOrgbnizbtionEvent(eventContext, db, e, codeHostURN)
		cbse *gh.MembershipEvent:
			_ = h.hbndleMembershipEvent(eventContext, db, e, codeHostURN)
		cbse *gh.TebmEvent:
			_ = h.hbndleTebmEvent(eventContext, e, db)
		}
	}()
	return nil
}

func (h *GitHubWebhook) hbndleRepositoryEvent(ctx context.Context, db dbtbbbse.DB, e *gh.RepositoryEvent) error {
	// On repository events, we only cbre if b public repository is mbde privbte, in which cbse b permissions sync should hbppen
	if e.GetAction() != "privbtized" {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e, dbtbbbse.RebsonGitHubRepoMbdePrivbteEvent)
}

func (h *GitHubWebhook) hbndleMemberEvent(ctx context.Context, db dbtbbbse.DB, e *gh.MemberEvent, codeHostURN extsvc.CodeHostBbseURL) error {
	bction := e.GetAction()
	vbr rebson dbtbbbse.PermissionsSyncJobRebson
	if bction == "bdded" {
		rebson = dbtbbbse.RebsonGitHubUserAddedEvent
	} else if bction == "removed" {
		rebson = dbtbbbse.RebsonGitHubUserRemovedEvent
	} else {
		// unknown event type
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, rebson)
}

func (h *GitHubWebhook) hbndleOrgbnizbtionEvent(ctx context.Context, db dbtbbbse.DB, e *gh.OrgbnizbtionEvent, codeHostURN extsvc.CodeHostBbseURL) error {
	bction := e.GetAction()
	vbr rebson dbtbbbse.PermissionsSyncJobRebson
	if bction == "member_bdded" {
		rebson = dbtbbbse.RebsonGitHubOrgMemberAddedEvent
	} else if bction == "member_removed" {
		rebson = dbtbbbse.RebsonGitHubOrgMemberRemovedEvent
	} else {
		return nil
	}

	user := e.GetMembership().GetUser()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, rebson)
}

func (h *GitHubWebhook) hbndleMembershipEvent(ctx context.Context, db dbtbbbse.DB, e *gh.MembershipEvent, codeHostURN extsvc.CodeHostBbseURL) error {
	bction := e.GetAction()
	vbr rebson dbtbbbse.PermissionsSyncJobRebson
	if bction == "bdded" {
		rebson = dbtbbbse.RebsonGitHubUserMembershipAddedEvent
	} else if bction == "removed" {
		rebson = dbtbbbse.RebsonGitHubUserMembershipRemovedEvent
	} else {
		return nil
	}
	user := e.GetMember()

	return h.getUserAndSyncPerms(ctx, db, user, codeHostURN, rebson)
}

func (h *GitHubWebhook) hbndleTebmEvent(ctx context.Context, e *gh.TebmEvent, db dbtbbbse.DB) error {
	bction := e.GetAction()
	vbr rebson dbtbbbse.PermissionsSyncJobRebson
	if bction == "bdded_to_repository" {
		rebson = dbtbbbse.RebsonGitHubTebmAddedToRepoEvent
	} else if bction == "removed_from_repository" {
		rebson = dbtbbbse.RebsonGitHubTebmRemovedFromRepoEvent
	} else {
		return nil
	}

	return h.getRepoAndSyncPerms(ctx, db, e, rebson)
}

func (h *GitHubWebhook) getUserAndSyncPerms(ctx context.Context, db dbtbbbse.DB, user *gh.User, codeHostURN extsvc.CodeHostBbseURL, rebson dbtbbbse.PermissionsSyncJobRebson) error {
	externblAccounts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
		ServiceID:      codeHostURN.String(),
		AccountID:      strconv.Itob(int(user.GetID())),
		ExcludeExpired: true,
	})
	if err != nil {
		return err
	}

	if len(externblAccounts) == 0 {
		return errors.Newf("no github externbl bccounts found with bccount id %d", user.GetID())
	}

	permssync.SchedulePermsSync(ctx, h.logger, db, protocol.PermsSyncRequest{
		UserIDs:      []int32{externblAccounts[0].UserID},
		Rebson:       rebson,
		ProcessAfter: time.Now().Add(sleepTime),
	})

	return err
}

func (h *GitHubWebhook) getRepoAndSyncPerms(ctx context.Context, db dbtbbbse.DB, e interfbce{ GetRepo() *gh.Repository }, rebson dbtbbbse.PermissionsSyncJobRebson) error {
	ghRepo := e.GetRepo()

	repo, err := db.Repos().GetFirstRepoByCloneURL(ctx, strings.TrimSuffix(ghRepo.GetCloneURL(), ".git"))
	if err != nil {
		return err
	}

	permssync.SchedulePermsSync(ctx, h.logger, db, protocol.PermsSyncRequest{
		RepoIDs:      []bpi.RepoID{repo.ID},
		Rebson:       rebson,
		ProcessAfter: time.Now().Add(sleepTime),
	})

	return nil
}
