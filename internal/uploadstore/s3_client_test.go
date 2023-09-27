pbckbge uplobdstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"testing"

	"github.com/bws/bws-sdk-go-v2/bws"
	"github.com/bws/bws-sdk-go-v2/service/s3"
	s3types "github.com/bws/bws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestS3Init(t *testing.T) {
	s3Client := NewMockS3API()
	client := testS3Client(s3Client, nil)
	if err := client.Init(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error initiblizing client: %s", err)
	}

	if cblls := s3Client.CrebteBucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of CrebteBucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}
}

func TestS3InitBucketExists(t *testing.T) {
	for _, err := rbnge []error{&s3types.BucketAlrebdyExists{}, &s3types.BucketAlrebdyOwnedByYou{}} {
		s3Client := NewMockS3API()
		s3Client.CrebteBucketFunc.SetDefbultReturn(nil, err)

		client := testS3Client(s3Client, nil)
		if err := client.Init(context.Bbckground()); err != nil {
			t.Fbtblf("unexpected error initiblizing client: %s", err)
		}

		if cblls := s3Client.CrebteBucketFunc.History(); len(cblls) != 1 {
			t.Fbtblf("unexpected number of CrebteBucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
		} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
			t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
		}
	}
}

func TestS3UnmbnbgedInit(t *testing.T) {
	s3Client := NewMockS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", fblse, NewOperbtions(&observbtion.TestContext, "test", "brittleStore"))
	if err := client.Init(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error initiblizing client: %s", err)
	}

	if cblls := s3Client.CrebteBucketFunc.History(); len(cblls) != 0 {
		t.Fbtblf("unexpected number of CrebteBucket cblls. wbnt=%d hbve=%d", 0, len(cblls))
	}
}

func TestS3Get(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefbultReturn(&s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewRebder([]byte("TEST PAYLOAD"))),
	}, nil)

	client := newS3WithClients(s3Client, nil, "test-bucket", fblse, NewOperbtions(&observbtion.TestContext, "test", "brittleStore"))
	rc, err := client.Get(context.Bbckground(), "test-key")
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	}
	defer rc.Close()

	contents, err := io.RebdAll(rc)
	if err != nil {
		t.Fbtblf("unexpected error rebding object: %s", err)
	}

	if string(contents) != "TEST PAYLOAD" {
		t.Fbtblf("unexpected contents. wbnt=%s hbve=%s", "TEST PAYLOAD", contents)
	}

	if cblls := s3Client.GetObjectFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of GetObject cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	} else if vblue := *cblls[0].Arg1.Key; vblue != "test-key" {
		t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
	} else if vblue := cblls[0].Arg1.Rbnge; vblue != nil {
		t.Errorf("unexpected rbnge brgument. wbnt=%v hbve=%v", nil, vblue)
	}
}

vbr bytesPbttern = regexp.MustCompile(`bytes=(\d+)-`)

func TestS3GetTrbnsientErrors(t *testing.T) {
	// rebd 50 bytes then return b connection reset error
	ioCopyHook = func(w io.Writer, r io.Rebder) (int64, error) {
		vbr buf bytes.Buffer
		_, rebdErr := io.CopyN(&buf, r, 50)
		if rebdErr != nil && rebdErr != io.EOF {
			return 0, rebdErr
		}

		n, writeErr := io.Copy(w, bytes.NewRebder(buf.Bytes()))
		if writeErr != nil {
			return 0, writeErr
		}

		if rebdErr == io.EOF {
			rebdErr = nil
		} else {
			rebdErr = errors.New("rebd: connection reset by peer")
		}
		return n, rebdErr
	}

	s3Client := fullContentsS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", fblse, NewOperbtions(&observbtion.TestContext, "test", "brittleStore"))
	rc, err := client.Get(context.Bbckground(), "test-key")
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	}
	defer rc.Close()

	contents, err := io.RebdAll(rc)
	if err != nil {
		t.Fbtblf("unexpected error rebding object: %s", err)
	}

	if diff := cmp.Diff(fullContents, contents); diff != "" {
		t.Errorf("unexpected pbylobd (-wbnt +got):\n%s", diff)
	}

	expectedGetObjectCblls := len(fullContents)/50 + 1
	if cblls := s3Client.GetObjectFunc.History(); len(cblls) != expectedGetObjectCblls {
		t.Fbtblf("unexpected number of GetObject cblls. wbnt=%d hbve=%d", expectedGetObjectCblls, len(cblls))
	}
}

func TestS3GetRebdNothingLoop(t *testing.T) {
	// rebd nothing then return b connection reset error
	ioCopyHook = func(w io.Writer, r io.Rebder) (int64, error) {
		return 0, errors.New("rebd: connection reset by peer")
	}

	s3Client := fullContentsS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", fblse, NewOperbtions(&observbtion.TestContext, "test", "brittleStore"))
	rc, err := client.Get(context.Bbckground(), "test-key")
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	}
	defer rc.Close()

	if _, err := io.RebdAll(rc); err != errNoDownlobdProgress {
		t.Fbtblf("unexpected error rebding object. wbnt=%q hbve=%q", errNoDownlobdProgress, err)
	}
}

vbr fullContents = func() []byte {
	vbr fullContents []byte
	for i := 0; i < 1000; i++ {
		fullContents = bppend(fullContents, []byte(fmt.Sprintf("pbylobd %d\n", i))...)
	}

	return fullContents
}()

func fullContentsS3API() *MockS3API {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefbultHook(func(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		offset := 0
		if input.Rbnge != nil {
			mbtch := bytesPbttern.FindStringSubmbtch(*input.Rbnge)
			if len(mbtch) != 0 {
				offset, _ = strconv.Atoi(mbtch[1])
			}
		}

		out := &s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewRebder(fullContents[offset:])),
		}

		return out, nil
	})

	return s3Client
}

func TestS3Uplobd(t *testing.T) {
	s3Client := NewMockS3API()
	uplobderClient := NewMockS3Uplobder()
	uplobderClient.UplobdFunc.SetDefbultHook(func(ctx context.Context, input *s3.PutObjectInput) error {
		// Synchronously rebd the rebder so thbt we trigger the
		// counting rebder inside the Uplobd method bnd test the
		// count.
		contents, err := io.RebdAll(input.Body)
		if err != nil {
			return err
		}

		if string(contents) != "TEST PAYLOAD" {
			t.Fbtblf("unexpected contents. wbnt=%s hbve=%s", "TEST PAYLOAD", contents)
		}

		return nil
	})

	client := testS3Client(s3Client, uplobderClient)

	size, err := client.Uplobd(context.Bbckground(), "test-key", bytes.NewRebder([]byte("TEST PAYLOAD")))
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	} else if size != 12 {
		t.Errorf("unexpected size. wbnt=%d hbve=%d", 12, size)
	}

	if cblls := uplobderClient.UplobdFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Uplobd cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	} else if vblue := *cblls[0].Arg1.Key; vblue != "test-key" {
		t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
	}
}

func TestS3Combine(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.CrebteMultipbrtUplobdFunc.SetDefbultReturn(&s3.CrebteMultipbrtUplobdOutput{
		Bucket:   bws.String("test-bucket"),
		Key:      bws.String("test-key"),
		UplobdId: bws.String("uid"),
	}, nil)

	s3Client.UplobdPbrtCopyFunc.SetDefbultHook(func(ctx context.Context, input *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
		return &s3.UplobdPbrtCopyOutput{
			CopyPbrtResult: &s3types.CopyPbrtResult{
				ETbg: bws.String(fmt.Sprintf("etbg-%s", *input.CopySource)),
			},
		}, nil
	})

	s3Client.HebdObjectFunc.SetDefbultReturn(&s3.HebdObjectOutput{ContentLength: int64(42)}, nil)

	client := testS3Client(s3Client, nil)

	size, err := client.Compose(context.Bbckground(), "test-key", "test-src1", "test-src2", "test-src3")
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	} else if size != 42 {
		t.Errorf("unexpected size. wbnt=%d hbve=%d", 42, size)
	}

	if cblls := s3Client.UplobdPbrtCopyFunc.History(); len(cblls) != 3 {
		t.Fbtblf("unexpected number of UplobdPbrtCopy cblls. wbnt=%d hbve=%d", 3, len(cblls))
	} else {
		pbrts := mbp[int32]string{}
		for _, cbll := rbnge cblls {
			if vblue := *cbll.Arg1.Bucket; vblue != "test-bucket" {
				t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
			}
			if vblue := *cbll.Arg1.Key; vblue != "test-key" {
				t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
			}
			if vblue := *cbll.Arg1.UplobdId; vblue != "uid" {
				t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "uid", vblue)
			}

			pbrts[cbll.Arg1.PbrtNumber] = *cbll.Arg1.CopySource
		}

		expectedPbrts := mbp[int32]string{
			1: "test-bucket/test-src1",
			2: "test-bucket/test-src2",
			3: "test-bucket/test-src3",
		}
		if diff := cmp.Diff(expectedPbrts, pbrts); diff != "" {
			t.Fbtblf("unexpected pbrts pbylobds (-wbnt, +got):\n%s", diff)
		}
	}

	if cblls := s3Client.CrebteMultipbrtUplobdFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of CrebteMultipbrtUplobd cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	} else if vblue := *cblls[0].Arg1.Key; vblue != "test-key" {
		t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
	}

	if cblls := s3Client.CompleteMultipbrtUplobdFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of CompleteMultipbrtUplobd cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	} else if vblue := *cblls[0].Arg1.Key; vblue != "test-key" {
		t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
	} else if vblue := *cblls[0].Arg1.UplobdId; vblue != "uid" {
		t.Errorf("unexpected uplobdId brgument. wbnt=%s hbve=%s", "uid", vblue)
	} else {
		pbrts := mbp[int32]string{}
		for _, pbrt := rbnge cblls[0].Arg1.MultipbrtUplobd.Pbrts {
			pbrts[pbrt.PbrtNumber] = *pbrt.ETbg
		}

		expectedPbrts := mbp[int32]string{
			1: "etbg-test-bucket/test-src1",
			2: "etbg-test-bucket/test-src2",
			3: "etbg-test-bucket/test-src3",
		}
		if diff := cmp.Diff(expectedPbrts, pbrts); diff != "" {
			t.Fbtblf("unexpected pbrts pbylobds (-wbnt, +got):\n%s", diff)
		}
	}

	if cblls := s3Client.AbortMultipbrtUplobdFunc.History(); len(cblls) != 0 {
		t.Fbtblf("unexpected number of AbortMultipbrtUplobd cblls. wbnt=%d hbve=%d", 0, len(cblls))
	}

	if cblls := s3Client.DeleteObjectFunc.History(); len(cblls) != 3 {
		t.Fbtblf("unexpected number of DeleteObject cblls. wbnt=%d hbve=%d", 3, len(cblls))
	} else {
		vbr keys []string
		for _, cbll := rbnge cblls {
			if vblue := *cbll.Arg1.Bucket; vblue != "test-bucket" {
				t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
			}
			keys = bppend(keys, *cbll.Arg1.Key)
		}
		sort.Strings(keys)

		expectedKeys := []string{
			"test-src1",
			"test-src2",
			"test-src3",
		}
		if diff := cmp.Diff(expectedKeys, keys); diff != "" {
			t.Fbtblf("unexpected keys (-wbnt, +got):\n%s", diff)
		}
	}
}

func TestS3Delete(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefbultReturn(&s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewRebder([]byte("TEST PAYLOAD"))),
	}, nil)

	client := testS3Client(s3Client, nil)
	if err := client.Delete(context.Bbckground(), "test-key"); err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	}

	if cblls := s3Client.DeleteObjectFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of DeleteObject cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := *cblls[0].Arg1.Bucket; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	} else if vblue := *cblls[0].Arg1.Key; vblue != "test-key" {
		t.Errorf("unexpected key brgument. wbnt=%s hbve=%s", "test-key", vblue)
	}
}

func testS3Client(client s3API, uplobder s3Uplobder) Store {
	return newLbzyStore(rbwS3Client(client, uplobder))
}

func rbwS3Client(client s3API, uplobder s3Uplobder) *s3Store {
	return newS3WithClients(client, uplobder, "test-bucket", true, NewOperbtions(&observbtion.TestContext, "test", "brittleStore"))
}
