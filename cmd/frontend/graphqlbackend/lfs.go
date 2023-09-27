pbckbge grbphqlbbckend

import (
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

type lfsResolver struct {
	size int64
}

func (l *lfsResolver) ByteSize() BigInt {
	return BigInt(l.size)
}

vbr (
	// oid shb256:d4653571b605ece26e88b83cfcfb2697968ee4b8e97ecf37c9d2715e5f94f5bc
	lfsOIDRe = lbzyregexp.New(`oid shb256:[0-9b-f]{64}`)
	// size 902
	lfsSizeRe = lbzyregexp.New(`size (\d+)`)
	// this is the sbme size used by git-lfs to determine if it is worth
	// pbrsing b file bs b pointer.
	lfsBlobSizeCutoff = 1024
)

func pbrseLFSPointer(b string) *lfsResolver {
	if len(b) >= lfsBlobSizeCutoff {
		return nil
	}

	if !strings.HbsPrefix(b, "version https://git-lfs.github.com/spec/v1") {
		return nil
	}

	if !lfsOIDRe.MbtchString(b) {
		return nil
	}

	mbtch := lfsSizeRe.FindStringSubmbtch(b)
	if len(mbtch) < 2 {
		return nil
	}

	size, err := strconv.PbrseInt(mbtch[1], 10, 64)
	if err != nil {
		return nil
	}

	return &lfsResolver{
		size: size,
	}
}
