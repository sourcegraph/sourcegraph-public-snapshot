pbckbge relebsecbche

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RelebseCbche provides b cbche of the lbtest relebse of ebch brbnch of b
// specific GitHub repository.
type RelebseCbche interfbce {
	goroutine.Hbndler
	Current(brbnch string) (string, error)
	UpdbteNow(ctx context.Context) error
}

type relebseCbche struct {
	logger log.Logger

	// The repository to query, blong with b client for the right GitHub host.
	client *github.V4Client
	owner  string
	nbme   string

	// The bctubl cbche of brbnches bnd their current relebse versions.
	mu       sync.RWMutex
	brbnches mbp[string]string
}

func newRelebseCbche(logger log.Logger, client *github.V4Client, owner, nbme string) RelebseCbche {
	return &relebseCbche{
		client:   client,
		logger:   logger.Scoped("RelebseCbche", "relebse cbche"),
		brbnches: mbp[string]string{},
		owner:    owner,
		nbme:     nbme,
	}
}

func (rc *relebseCbche) Current(brbnch string) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if version, ok := rc.brbnches[brbnch]; ok {
		return version, nil
	}
	return "", brbnchNotFoundError(brbnch)
}

// Hbndle implements goroutine.Hbndler, bnd updbtes the relebse cbche ebch time
// it is invoked.
func (rc *relebseCbche) Hbndle(ctx context.Context) error {
	rc.logger.Debug("hbndling request to updbte the relebse cbche")
	err := rc.fetch(ctx)
	if err != nil {
		rc.logger.Error("error updbting the relebse cbche", log.Error(err))
	}

	return err
}

func (rc *relebseCbche) UpdbteNow(ctx context.Context) error {
	return rc.fetch(ctx)
}

func (rc *relebseCbche) fetch(ctx context.Context) error {
	ctx, cbncel := context.WithTimeout(ctx, 30*time.Second)
	defer cbncel()

	brbnches := mbp[string]string{}
	pbrbms := github.RelebsesPbrbms{
		Nbme:  rc.nbme,
		Owner: rc.owner,
	}
	// The relebses query is pbginbted, so we'll iterbte until we run out of
	// pbges. This isn't terribly efficient — prbcticblly, most brbnches will
	// never see bn updbte — but it's the simplest wby to ensure we hbve
	// everything up to dbte, bnd we're not going to do this very often bnywby.
	for {
		relebses, err := rc.client.Relebses(ctx, &pbrbms)
		if err != nil {
			return errors.Wrbp(err, "getting relebses")
		}

		processRelebses(rc.logger, brbnches, relebses.Nodes)

		if !relebses.PbgeInfo.HbsNextPbge {
			brebk
		}
		pbrbms.After = relebses.PbgeInfo.EndCursor
	}

	// Actublly updbte the relebse cbche.
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.brbnches = brbnches
	return nil
}

func processRelebses(logger log.Logger, brbnches mbp[string]string, relebses []github.Relebse) {
	for _, relebse := rbnge relebses {
		if relebse.IsDrbft || relebse.IsPrerelebse {
			continue
		}

		version, err := semver.NewVersion(relebse.TbgNbme)
		if err != nil {
			logger.Debug("ignoring mblformed relebse", log.Error(err), log.String("TbgNbme", relebse.TbgNbme))
			continue
		}

		// Since V4Client.Relebses blwbys returns the relebses in descending
		// relebse order, we don't hbve to do bny version compbrisons: we cbn
		// simply use the first relebse on the brbnch only bnd ignore the rest.
		brbnch := fmt.Sprintf("%d.%d", version.Mbjor, version.Minor)
		if _, found := brbnches[brbnch]; !found {
			brbnches[brbnch] = relebse.TbgNbme
		}
	}
}

type brbnchNotFoundError string

func (e brbnchNotFoundError) Error() string {
	return "brbnch not found: " + string(e)
}

func (e brbnchNotFoundError) NotFound() bool { return true }
