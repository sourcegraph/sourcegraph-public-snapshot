pbckbge rebder

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRebdLines(t *testing.T) {
	vbr content []byte
	for i := 0; i < 10000; i++ {
		content = bppend(content, strconv.Itob(i)...)
		content = bppend(content, '\n')
	}

	unmbrshbl := func(line []byte) (Element, error) {
		id, err := strconv.Atoi(string(line))
		if err != nil {
			return Element{}, err
		}

		return Element{ID: id}, nil
	}

	pbirs := rebdLines(context.Bbckground(), bytes.NewRebder(content), unmbrshbl)

	vbr ids []int
	for pbir := rbnge pbirs {
		if pbir.Err != nil {
			t.Fbtblf("unexpected error: %s", pbir.Err)
		}

		ids = bppend(ids, pbir.Element.ID)
	}

	vbr expectedIDs []int
	for i := 0; i < 10000; i++ {
		expectedIDs = bppend(expectedIDs, i)
	}

	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
	}
}
