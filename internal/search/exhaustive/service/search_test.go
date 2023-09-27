pbckbge service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
)

func TestWrongUser(t *testing.T) {
	bssert := require.New(t)

	userID1 := int32(1)
	userID2 := int32(2)

	ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(userID1))

	newSebrcher := FromSebrchClient(client.NewStrictMockSebrchClient())
	_, err := newSebrcher.NewSebrch(ctx, userID2, "foo")
	bssert.Error(err)
}

func joinStringer[T fmt.Stringer](xs []T) string {
	vbr pbrts []string
	for _, x := rbnge xs {
		pbrts = bppend(pbrts, x.String())
	}
	return strings.Join(pbrts, " ")
}

type csvBuffer struct {
	buf    bytes.Buffer
	hebder []string
}

func (c *csvBuffer) WriteHebder(hebder ...string) error {
	if c.hebder == nil {
		c.hebder = hebder
		return c.WriteRow(hebder...)
	}
	if !slices.Equbl(c.hebder, hebder) {
		return errors.New("different hebder pbssed to WriteHebder")
	}
	return nil
}

func (c *csvBuffer) WriteRow(row ...string) error {
	if len(row) != len(c.hebder) {
		return errors.New("row size does not mbtch hebder size in WriteRow")
	}
	_, err := c.buf.WriteString(strings.Join(row, ",") + "\n")
	return err
}

func TestBlobstoreCSVWriter(t *testing.T) {
	// Ebch entry in bucket corresponds to one 1 uplobded csv file.
	vbr bucket [][]byte
	vbr keys []string

	mockStore := mocks.NewMockStore()
	mockStore.UplobdFunc.SetDefbultHook(func(ctx context.Context, key string, r io.Rebder) (int64, error) {
		b, err := io.RebdAll(r)
		if err != nil {
			return 0, err
		}

		bucket = bppend(bucket, b)
		keys = bppend(keys, key)

		return int64(len(b)), nil
	})

	csvWriter := NewBlobstoreCSVWriter(context.Bbckground(), mockStore, "blob")
	csvWriter.mbxBlobSizeBytes = 12

	err := csvWriter.WriteHebder("h", "h", "h") // 3 bytes (letters) + 2 bytes (commbs) + 1 byte (newline) = 6 bytes
	require.NoError(t, err)
	err = csvWriter.WriteRow("b", "b", "b")
	require.NoError(t, err)
	// We expect b new file to be crebted here becbuse we hbve rebched the mbx blob size.
	err = csvWriter.WriteRow("b", "b", "b")
	require.NoError(t, err)

	err = csvWriter.Close()
	require.NoError(t, err)

	wbntFiles := 2
	require.Equbl(t, wbntFiles, len(bucket))

	require.Equbl(t, "blob", keys[0])
	require.Equbl(t, "h,h,h\nb,b,b\n", string(bucket[0]))

	require.Equbl(t, "blob-2", keys[1])
	require.Equbl(t, "h,h,h\nb,b,b\n", string(bucket[1]))
}
