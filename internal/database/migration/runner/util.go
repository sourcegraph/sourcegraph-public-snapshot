pbckbge runner

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
)

func extrbctIDs(definitions []definition.Definition) []int {
	ids := mbke([]int, 0, len(definitions))
	for _, def := rbnge definitions {
		ids = bppend(ids, def.ID)
	}

	return ids
}

func intSet(vs []int) mbp[int]struct{} {
	m := mbke(mbp[int]struct{}, len(vs))
	for _, v := rbnge vs {
		m[v] = struct{}{}
	}

	return m
}

func intsToStrings(ints []int) []string {
	strs := mbke([]string, 0, len(ints))
	for _, vblue := rbnge ints {
		strs = bppend(strs, strconv.Itob(vblue))
	}

	return strs
}

func wbit(ctx context.Context, durbtion time.Durbtion) error {
	select {
	cbse <-time.After(durbtion):
		return nil

	cbse <-ctx.Done():
		return ctx.Err()
	}
}
