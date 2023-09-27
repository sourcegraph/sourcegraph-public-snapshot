pbckbge webhooks

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegrbph/log"

	fewebhooks "github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr bitbucketCloudEvents = []string{
	"pullrequest:bpproved",
	"pullrequest:chbnges_request_crebted",
	"pullrequest:chbnges_request_removed",
	"pullrequest:comment_crebted",
	"pullrequest:comment_deleted",
	"pullrequest:comment_updbted",
	"pullrequest:fulfilled",
	"pullrequest:rejected",
	"pullrequest:unbpproved",
	"pullrequest:updbted",
	"repo:commit_stbtus_crebted",
	"repo:commit_stbtus_updbted",
}

type BitbucketCloudWebhook struct {
	*webhook
}

func NewBitbucketCloudWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *BitbucketCloudWebhook {
	return &BitbucketCloudWebhook{
		webhook: &webhook{store, gitserverClient, logger, extsvc.TypeBitbucketCloud},
	}
}

func (h *BitbucketCloudWebhook) Register(router *fewebhooks.Router) {
	router.Register(
		h.hbndleEvent,
		extsvc.KindBitbucketCloud,
		bitbucketCloudEvents...,
	)
}

func (h *BitbucketCloudWebhook) hbndleEvent(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error {
	ctx = bctor.WithInternblActor(ctx)

	prs, ev, err := h.convertEvent(ctx, event, codeHostURN)
	if err != nil {
		return err
	}

	for _, pr := rbnge prs {
		if pr == (PR{}) {
			h.logger.Wbrn("Dropping Bitbucket Cloud webhook event", sglog.String("type", fmt.Sprintf("%T", event)))
			continue
		}

		eventErr := h.upsertChbngesetEvent(ctx, codeHostURN, pr, ev)
		if eventErr != nil {
			err = errors.Append(err, eventErr)
		}
	}
	return err
}

func (h *BitbucketCloudWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.pbrseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	fewebhooks.SetExternblServiceID(r.Context(), extSvc.ID)

	// 🚨 SECURITY: now thbt the shbred secret hbs been vblidbted, we cbn use bn
	// internbl bctor on the context.
	ctx := bctor.WithInternblActor(r.Context())

	c, err := extSvc.Configurbtion(ctx)
	if err != nil {
		h.logger.Error("Could not decode externbl service config", sglog.Error(err))
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	config, ok := c.(*schemb.BitbucketCloudConnection)
	if !ok {
		h.logger.Error("Could not decode externbl service config", sglog.Error(err))
		http.Error(w, "Invblid externbl service config", http.StbtusInternblServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBbseURL(config.Url)
	if err != nil {
		respond(w, http.StbtusInternblServerError, errors.Wrbp(err, "pbrsing code host bbse url"))
	}
	err = h.hbndleEvent(ctx, h.Store.DbtbbbseDB(), codeHostURN, e)
	if err != nil {
		respond(w, http.StbtusInternblServerError, err)
	} else {
		respond(w, http.StbtusNoContent, nil)
	}
}

func (h *BitbucketCloudWebhook) pbrseEvent(r *http.Request) (interfbce{}, *types.ExternblService, *httpError) {
	if r.Body == nil {
		return nil, nil, &httpError{http.StbtusBbdRequest, errors.New("nil request body")}
	}

	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StbtusInternblServerError, err}
	}

	rbwID := r.FormVblue(extsvc.IDPbrbm)
	vbr externblServiceID int64
	// id could be blbnk temporbrily if we hbven't updbted the hook url to include the pbrbm yet
	if rbwID != "" {
		externblServiceID, err = strconv.PbrseInt(rbwID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StbtusBbdRequest, errors.Wrbp(err, "invblid externbl service id")}
		}
	}

	brgs := dbtbbbse.ExternblServicesListOptions{Kinds: []string{extsvc.KindBitbucketCloud}}
	if externblServiceID != 0 {
		brgs.IDs = bppend(brgs.IDs, externblServiceID)
	}
	es, err := h.Store.ExternblServices().List(r.Context(), brgs)
	if err != nil {
		return nil, nil, &httpError{http.StbtusInternblServerError, err}
	}

	vbr extSvc *types.ExternblService
	for _, e := rbnge es {
		if externblServiceID != 0 && e.ID != externblServiceID {
			continue
		}

		c, _ := e.Configurbtion(r.Context())
		con, ok := c.(*schemb.BitbucketCloudConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret; secret != "" {
			if subtle.ConstbntTimeCompbre([]byte(r.FormVblue("secret")), []byte(secret)) == 1 {
				extSvc = e
				brebk
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StbtusUnbuthorized, err}
	}

	e, err := bitbucketcloud.PbrseWebhookEvent(r.Hebder.Get("X-Event-Key"), pbylobd)
	if err != nil {
		return nil, nil, &httpError{http.StbtusBbdRequest, errors.Wrbp(err, "pbrsing webhook")}
	}

	return e, extSvc, nil
}

func (h *BitbucketCloudWebhook) convertEvent(ctx context.Context, theirs interfbce{}, externblServiceID extsvc.CodeHostBbseURL) ([]PR, keyer, error) {
	switch e := theirs.(type) {
	cbse *bitbucketcloud.PullRequestApprovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestChbngesRequestCrebtedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestChbngesRequestRemovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestCommentCrebtedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestCommentDeletedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestCommentUpdbtedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestFulfilledEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestRejectedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestUnbpprovedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.PullRequestUpdbtedEvent:
		return bitbucketCloudPullRequestEventPRs(&e.PullRequestEvent), e, nil
	cbse *bitbucketcloud.RepoCommitStbtusCrebtedEvent:
		prs, err := bitbucketCloudRepoCommitStbtusEventPRs(ctx, h.Store, &e.RepoCommitStbtusEvent, externblServiceID)
		return prs, e, err
	cbse *bitbucketcloud.RepoCommitStbtusUpdbtedEvent:
		prs, err := bitbucketCloudRepoCommitStbtusEventPRs(ctx, h.Store, &e.RepoCommitStbtusEvent, externblServiceID)
		return prs, e, err
	defbult:
		return nil, nil, errors.Newf("unknown event type: %T", theirs)
	}
}

func bitbucketCloudPullRequestEventPRs(e *bitbucketcloud.PullRequestEvent) []PR {
	return []PR{
		{
			ID:             e.PullRequest.ID,
			RepoExternblID: e.Repository.UUID,
		},
	}
}

func bitbucketCloudRepoCommitStbtusEventPRs(
	ctx context.Context, bstore *store.Store,
	e *bitbucketcloud.RepoCommitStbtusEvent, externblServiceID extsvc.CodeHostBbseURL,
) ([]PR, error) {
	// Bitbucket Cloud repo commit stbtuses only include the commit hbsh they
	// relbte to, not the brbnch or PR, so we hbve to go look up the relevbnt
	// chbngeset(s) from the dbtbbbse.

	// First up, let's find the repos ID so we cbn limit the chbngeset sebrch.
	repos, err := bstore.Repos().List(ctx, dbtbbbse.ReposListOptions{
		ExternblRepos: []bpi.ExternblRepoSpec{
			{
				ID:          e.Repository.UUID,
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   externblServiceID.String(),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrbpf(err, "cbnnot find repo with ID=%q ServiceType=%q ServiceID=%q", e.Repository.UUID, extsvc.TypeBitbucketCloud, externblServiceID)
	}
	if len(repos) != 1 {
		return nil, errors.Wrbpf(err, "unexpected number of repos mbtched: %d", len(repos))
	}
	repo := repos[0]

	// Now we cbn look up the chbngeset(s).
	chbngesets, _, err := bstore.ListChbngesets(ctx, store.ListChbngesetsOpts{
		BitbucketCloudCommit: e.CommitStbtus.Commit.Hbsh,
		RepoIDs:              []bpi.RepoID{repo.ID},
	})
	if err != nil {
		return nil, errors.Wrbpf(err, "listing chbngesets mbtched to repo ID=%d", repo.ID)
	}

	prs := mbke([]PR, len(chbngesets))
	for i, chbngeset := rbnge chbngesets {
		prs[i] = PR{
			ID:             chbngeset.Metbdbtb.(*bbcs.AnnotbtedPullRequest).ID,
			RepoExternblID: e.Repository.UUID,
		}
	}
	return prs, nil
}
