package graphqlbackend

import (
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

type lfsResolver struct {
	// TODO what if file is bigger than 4gb? This seems likely for LFS. Do we
	// need to return a float?
	size int32
}

func (l *lfsResolver) ByteSize() int32 {
	return l.size
}

var (
	// oid sha256:d4653571a605ece26e88b83cfcfa2697968ee4b8e97ecf37c9d2715e5f94f5ac
	lfsOIDRe = lazyregexp.New(`oid sha256:[0-9a-f]{64}`)
	// size 902
	lfsSizeRe = lazyregexp.New(`size (\d+)`)
	// this is the same size used by git-lfs to determine if it is worth
	// parsing a file as a pointer.
	lfsBlobSizeCutoff = 1024
)

func parseLFSPointer(b string) *lfsResolver {
	if len(b) >= lfsBlobSizeCutoff {
		return nil
	}

	if !strings.HasPrefix(b, "version https://git-lfs.github.com/spec/v1") {
		return nil
	}

	if !lfsOIDRe.MatchString(b) {
		return nil
	}

	match := lfsSizeRe.FindStringSubmatch(b)
	if len(match) < 2 {
		return nil
	}

	size, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		return nil
	}

	return &lfsResolver{
		size: int32(size),
	}
}
