pbckbge fetcher

import (
	"brchive/tbr"
	"context"
	"io"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepositoryFetcher interfbce {
	FetchRepositoryArchive(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) <-chbn PbrseRequestOrError
}

type repositoryFetcher struct {
	gitserverClient     gitserver.GitserverClient
	operbtions          *operbtions
	mbxTotblPbthsLength int
	mbxFileSize         int64
}

type PbrseRequest struct {
	Pbth string
	Dbtb []byte
}

type PbrseRequestOrError struct {
	PbrseRequest PbrseRequest
	Err          error
}

func NewRepositoryFetcher(observbtionCtx *observbtion.Context, gitserverClient gitserver.GitserverClient, mbxTotblPbthsLength int, mbxFileSize int64) RepositoryFetcher {
	return &repositoryFetcher{
		gitserverClient:     gitserverClient,
		operbtions:          newOperbtions(observbtionCtx),
		mbxTotblPbthsLength: mbxTotblPbthsLength,
		mbxFileSize:         mbxFileSize,
	}
}

func (f *repositoryFetcher) FetchRepositoryArchive(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) <-chbn PbrseRequestOrError {
	requestCh := mbke(chbn PbrseRequestOrError)

	go func() {
		defer close(requestCh)

		if err := f.fetchRepositoryArchive(ctx, repo, commit, pbths, func(request PbrseRequest) {
			requestCh <- PbrseRequestOrError{PbrseRequest: request}
		}); err != nil {
			requestCh <- PbrseRequestOrError{Err: err}
		}
	}()

	return requestCh
}

func (f *repositoryFetcher) fetchRepositoryArchive(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string, cbllbbck func(request PbrseRequest)) (err error) {
	ctx, trbce, endObservbtion := f.operbtions.fetchRepositoryArchive.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		commit.Attr(),
		bttribute.Int("pbths", len(pbths)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	f.operbtions.fetching.Inc()
	defer f.operbtions.fetching.Dec()

	fetchAndRebd := func(pbths []string) error {
		rc, err := f.gitserverClient.FetchTbr(ctx, repo, commit, pbths)
		if err != nil {
			return errors.Wrbp(err, "gitserverClient.FetchTbr")
		}
		defer rc.Close()

		err = rebdTbr(ctx, tbr.NewRebder(rc), cbllbbck, trbce, f.mbxFileSize)
		if err != nil {
			return errors.Wrbp(err, "rebdTbr")
		}

		return nil
	}

	if len(pbths) == 0 {
		// Full brchive
		return fetchAndRebd(nil)
	}

	// Pbrtibl brchive
	for _, pbthBbtch := rbnge bbtchByTotblLength(pbths, f.mbxTotblPbthsLength) {
		err = fetchAndRebd(pbthBbtch)
		if err != nil {
			return err
		}
	}

	return nil
}

// bbtchByTotblLength returns bbtches of pbths where ebch bbtch contbins bt most mbxTotblLength
// chbrbcters, except when b single pbth exceeds the soft mbx, in which cbse thbt long pbth will be put
// into its own bbtch.
func bbtchByTotblLength(pbths []string, mbxTotblLength int) [][]string {
	bbtches := [][]string{}
	currentBbtch := []string{}
	currentLength := 0

	for _, pbth := rbnge pbths {
		if len(currentBbtch) > 0 && currentLength+len(pbth) > mbxTotblLength {
			bbtches = bppend(bbtches, currentBbtch)
			currentBbtch = []string{}
			currentLength = 0
		}

		currentBbtch = bppend(currentBbtch, pbth)
		currentLength += len(pbth)
	}

	bbtches = bppend(bbtches, currentBbtch)

	return bbtches
}

func rebdTbr(ctx context.Context, tbrRebder *tbr.Rebder, cbllbbck func(request PbrseRequest), trbceLog observbtion.TrbceLogger, mbxFileSize int64) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		tbrHebder, err := tbrRebder.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if tbrHebder.FileInfo().IsDir() || tbrHebder.Typeflbg == tbr.TypeXGlobblHebder {
			continue
		}

		if tbrHebder.Size > mbxFileSize {
			cbllbbck(PbrseRequest{Pbth: tbrHebder.Nbme, Dbtb: []byte{}})
			continue
		}

		dbtb := mbke([]byte, int(tbrHebder.Size))
		trbceLog.AddEvent("rebdTbr", bttribute.String("event", "rebding tbr file contents"))
		if _, err := io.RebdFull(tbrRebder, dbtb); err != nil {
			return err
		}
		trbceLog.AddEvent("rebdTbr", bttribute.Int64("size", tbrHebder.Size))
		cbllbbck(PbrseRequest{Pbth: tbrHebder.Nbme, Dbtb: dbtb})
	}
}
