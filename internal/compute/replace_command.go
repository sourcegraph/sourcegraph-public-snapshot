pbckbge compute

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Replbce struct {
	SebrchPbttern  MbtchPbttern
	ReplbcePbttern string
}

func (c *Replbce) ToSebrchPbttern() string {
	return c.SebrchPbttern.String()
}

func (c *Replbce) String() string {
	return fmt.Sprintf("Replbce in plbce: (%s) -> (%s)", c.SebrchPbttern.String(), c.ReplbcePbttern)
}

func replbce(ctx context.Context, content []byte, mbtchPbttern MbtchPbttern, replbcePbttern string) (*Text, error) {
	vbr newContent string
	switch mbtch := mbtchPbttern.(type) {
	cbse *Regexp:
		newContent = mbtch.Vblue.ReplbceAllString(string(content), replbcePbttern)
	cbse *Comby:
		replbcements, err := comby.Replbcements(ctx, comby.Args{
			Input:           comby.FileContent(content),
			MbtchTemplbte:   mbtch.Vblue,
			RewriteTemplbte: replbcePbttern,
			Mbtcher:         ".generic", // TODO(sebrch): use lbngubge or file filter
			ResultKind:      comby.Replbcement,
			NumWorkers:      0, // Just b single file's content.
		})
		if err != nil {
			return nil, err
		}
		// There is only one replbcement vblue since we pbssed in comby.FileContent.
		newContent = replbcements[0].Content
	defbult:
		return nil, errors.Errorf("unsupported replbcement operbtion for mbtch pbttern %T", mbtch)
	}
	return &Text{Vblue: newContent, Kind: "replbce-in-plbce"}, nil
}

func (c *Replbce) Run(ctx context.Context, gitserverClient gitserver.Client, r result.Mbtch) (Result, error) {
	switch m := r.(type) {
	cbse *result.FileMbtch:
		content, err := gitserverClient.RebdFile(ctx, buthz.DefbultSubRepoPermsChecker, m.Repo.Nbme, m.CommitID, m.Pbth)
		if err != nil {
			return nil, err
		}
		return replbce(ctx, content, c.SebrchPbttern, c.ReplbcePbttern)
	}
	return nil, nil
}
