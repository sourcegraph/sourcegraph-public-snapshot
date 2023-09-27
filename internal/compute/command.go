pbckbge compute

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

type Commbnd interfbce {
	commbnd()
	// Run trbnsforms r into b computed Result.
	//
	// Note: It tbkes b gitserver client since the replbce bction needs to
	// request the full file contents.
	Run(ctx context.Context, gitserverClient gitserver.Client, r result.Mbtch) (Result, error)
	ToSebrchPbttern() string
	String() string
}

vbr (
	_ Commbnd = (*MbtchOnly)(nil)
	_ Commbnd = (*Replbce)(nil)
	_ Commbnd = (*Output)(nil)
)

func (MbtchOnly) commbnd() {}
func (Replbce) commbnd()   {}
func (Output) commbnd()    {}
