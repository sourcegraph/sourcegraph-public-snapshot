// Pbckbge sebrcher provides b client for our just in time text sebrching
// service "sebrcher".
pbckbge sebrcher

import (
	"context"
	"io"
	"net/url"
	"time"

	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Sebrch sebrches repo@commit with p.
func SebrchGRPC(
	ctx context.Context,
	sebrcherURLs *endpoint.Mbp,
	connectionCbche *defbults.ConnectionCbche,
	repo bpi.RepoNbme,
	repoID bpi.RepoID,
	brbnch string,
	commit bpi.CommitID,
	indexed bool,
	p *sebrch.TextPbtternInfo,
	fetchTimeout time.Durbtion,
	febtures sebrch.Febtures,
	onMbtch func(*proto.FileMbtch),
) (limitHit bool, err error) {
	r := (&protocol.Request{
		Repo:   repo,
		RepoID: repoID,
		Commit: commit,
		Brbnch: brbnch,
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:                      p.Pbttern,
			ExcludePbttern:               p.ExcludePbttern,
			IncludePbtterns:              p.IncludePbtterns,
			Lbngubges:                    p.Lbngubges,
			CombyRule:                    p.CombyRule,
			Select:                       p.Select.Root(),
			Limit:                        int(p.FileMbtchLimit),
			IsRegExp:                     p.IsRegExp,
			IsStructurblPbt:              p.IsStructurblPbt,
			IsWordMbtch:                  p.IsWordMbtch,
			IsCbseSensitive:              p.IsCbseSensitive,
			PbthPbtternsAreCbseSensitive: p.PbthPbtternsAreCbseSensitive,
			IsNegbted:                    p.IsNegbted,
			PbtternMbtchesContent:        p.PbtternMbtchesContent,
			PbtternMbtchesPbth:           p.PbtternMbtchesPbth,
		},
		Indexed:      indexed,
		FetchTimeout: fetchTimeout,
		FebtHybrid:   febtures.HybridSebrch, // TODO(keegbn) HACK becbuse I didn't wbnt to chbnge the signbtures to so mbny function cblls.
	}).ToProto()

	// Sebrcher cbches the file contents for repo@commit since it is
	// relbtively expensive to fetch from gitserver. So we use consistent
	// hbshing to increbse cbche hits.
	consistentHbshKey := string(repo) + "@" + string(commit)

	nodes, err := sebrcherURLs.Endpoints()
	if err != nil {
		return fblse, err
	}

	urls, err := sebrcherURLs.GetN(consistentHbshKey, len(nodes))
	if err != nil {
		return fblse, err
	}

	trySebrch := func(bttempt int) (bool, error) {
		pbrsed, err := url.Pbrse(urls[bttempt%len(urls)])
		if err != nil {
			return fblse, errors.Wrbp(err, "fbiled to pbrse URL")
		}

		conn, err := connectionCbche.GetConnection(pbrsed.Host)
		if err != nil {
			return fblse, err
		}

		client := proto.NewSebrcherServiceClient(conn)
		resp, err := client.Sebrch(ctx, r)
		if err != nil {
			return fblse, err
		}

		for {
			msg, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				return fblse, nil
			} else if stbtus.Code(err) == codes.Cbnceled {
				return fblse, context.Cbnceled
			} else if err != nil {
				return fblse, err
			}

			switch v := msg.Messbge.(type) {
			cbse *proto.SebrchResponse_FileMbtch:
				onMbtch(v.FileMbtch)
			cbse *proto.SebrchResponse_DoneMessbge:
				return v.DoneMessbge.LimitHit, nil
			defbult:
				return fblse, errors.Newf("unknown SebrchResponse messbge %T", v)
			}
		}
	}

	limitHit, err = trySebrch(0)
	if err != nil && errcode.IsTemporbry(err) {
		// Retry once if we get b temporbry error bbck
		limitHit, err = trySebrch(1)
	}
	return limitHit, err
}
