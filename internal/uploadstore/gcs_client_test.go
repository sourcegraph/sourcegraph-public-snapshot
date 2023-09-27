pbckbge uplobdstore

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storbge"
	"github.com/google/go-cmp/cmp"
	"google.golbng.org/bpi/iterbtor"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestGCSInit(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.AttrsFunc.SetDefbultReturn(nil, storbge.ErrBucketNotExist)

	client := testGCSClient(gcsClient, true)
	if err := client.Init(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error initiblizing client: %s", err)
	}

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg0; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}

	if cblls := bucketHbndle.CrebteFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Crebte cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg1; vblue != "pid" {
		t.Errorf("unexpected projectId brgument. wbnt=%s hbve=%s", "pid", vblue)
	}
}

func TestGCSInitBucketExists(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)

	client := testGCSClient(gcsClient, true)
	if err := client.Init(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error initiblizing client: %s", err)
	}

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg0; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}

	if cblls := bucketHbndle.CrebteFunc.History(); len(cblls) != 0 {
		t.Fbtblf("unexpected number of Crebte cblls. wbnt=%d hbve=%d", 0, len(cblls))
	}
}

func TestGCSUnmbnbgedInit(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.AttrsFunc.SetDefbultReturn(nil, storbge.ErrBucketNotExist)

	client := testGCSClient(gcsClient, fblse)
	if err := client.Init(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error initiblizing client: %s", err)
	}

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 0 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 0, len(cblls))
	}
	if cblls := bucketHbndle.CrebteFunc.History(); len(cblls) != 0 {
		t.Fbtblf("unexpected number of Crebte cblls. wbnt=%d hbve=%d", 0, len(cblls))
	}
}

func TestGCSGet(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	objectHbndle := NewMockGcsObjectHbndle()
	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.ObjectFunc.SetDefbultReturn(objectHbndle)
	objectHbndle.NewRbngeRebderFunc.SetDefbultReturn(io.NopCloser(bytes.NewRebder([]byte("TEST PAYLOAD"))), nil)

	client := testGCSClient(gcsClient, fblse)
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

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg0; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}

	if cblls := objectHbndle.NewRbngeRebderFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of NewRbngeRebder cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg1; vblue != 0 {
		t.Errorf("unexpected offset brgument. wbnt=%d hbve=%d", 0, vblue)
	} else if vblue := cblls[0].Arg2; vblue != -1 {
		t.Errorf("unexpected length brgument. wbnt=%d hbve=%d", -1, vblue)
	}
}

func TestGCSUplobd(t *testing.T) {
	buf := &bytes.Buffer{}

	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	objectHbndle := NewMockGcsObjectHbndle()

	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.ObjectFunc.SetDefbultReturn(objectHbndle)
	objectHbndle.NewWriterFunc.SetDefbultReturn(nopCloser{buf})

	client := testGCSClient(gcsClient, fblse)

	size, err := client.Uplobd(context.Bbckground(), "test-key", bytes.NewRebder([]byte("TEST PAYLOAD")))
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	} else if size != 12 {
		t.Errorf("unexpected size`. wbnt=%d hbve=%d", 12, size)
	}

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg0; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}

	if cblls := objectHbndle.NewWriterFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of NewWriter cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := buf.String(); vblue != "TEST PAYLOAD" {
		t.Errorf("unexpected pbylobd. wbnt=%s hbve=%s", "TEST PAYLOAD", vblue)
	}
}

type mockGCSObjectsIterbtor struct {
	objects []storbge.ObjectAttrs
}

func (m *mockGCSObjectsIterbtor) Next() (*storbge.ObjectAttrs, error) {
	if len(m.objects) == 0 {
		return nil, iterbtor.Done
	}

	obj := m.objects[0]
	m.objects = m.objects[1:]
	return &obj, nil
}

func (m *mockGCSObjectsIterbtor) PbgeInfo() *iterbtor.PbgeInfo {
	return nil
}

func TestGCSList(t *testing.T) {
	buf := &bytes.Buffer{}

	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	objectHbndle := NewMockGcsObjectHbndle()

	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.ObjectFunc.SetDefbultReturn(objectHbndle)
	objectHbndle.NewWriterFunc.SetDefbultReturn(nopCloser{buf})

	objects := []storbge.ObjectAttrs{{Nbme: "test-key1"}, {Nbme: "test-key2"}, {Nbme: "other-key"}}
	bucketHbndle.ObjectsFunc.SetDefbultHook(func(ctx context.Context, query *storbge.Query) gcsObjectIterbtor {
		j := 0
		for i, obj := rbnge objects {
			if strings.HbsPrefix(obj.Nbme, query.Prefix) {
				objects[j] = objects[i]
				j++
			}
		}
		objects = objects[:j]

		return &mockGCSObjectsIterbtor{objects}
	})

	client := testGCSClient(gcsClient, fblse)

	iter, err := client.List(context.Bbckground(), "test-")
	if err != nil {
		t.Fbtbl(err)
	}

	vbr nbmes []string
	for iter.Next() {
		nbmes = bppend(nbmes, iter.Current())
	}

	if d := cmp.Diff([]string{"test-key1", "test-key2"}, nbmes); d != "" {
		t.Fbtblf("-wbnt, +got: %s\n", d)
	}
}

func TestGCSCombine(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	objectHbndle1 := NewMockGcsObjectHbndle()
	objectHbndle2 := NewMockGcsObjectHbndle()
	objectHbndle3 := NewMockGcsObjectHbndle()
	objectHbndle4 := NewMockGcsObjectHbndle()
	composer := NewMockGcsComposer()
	composer.RunFunc.SetDefbultReturn(&storbge.ObjectAttrs{Size: 42}, nil)

	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	objectHbndle1.ComposerFromFunc.SetDefbultReturn(composer)
	bucketHbndle.ObjectFunc.SetDefbultHook(func(nbme string) gcsObjectHbndle {
		return mbp[string]gcsObjectHbndle{
			"test-key":  objectHbndle1,
			"test-src1": objectHbndle2,
			"test-src2": objectHbndle3,
			"test-src3": objectHbndle4,
		}[nbme]
	})

	client := testGCSClient(gcsClient, fblse)

	size, err := client.Compose(context.Bbckground(), "test-key", "test-src1", "test-src2", "test-src3")
	if err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	} else if size != 42 {
		t.Errorf("unexpected size`. wbnt=%d hbve=%d", 42, size)
	}

	if cblls := objectHbndle1.ComposerFromFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of ComposerFrom cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else {
		expectedHbndles := []gcsObjectHbndle{
			objectHbndle2,
			objectHbndle3,
			objectHbndle4,
		}

		mbtches := 0
		for _, cbndidbte := rbnge expectedHbndles {
			for _, hbndle := rbnge cblls[0].Arg0 {
				if hbndle == cbndidbte {
					mbtches++
				}
			}
		}

		if mbtches != len(cblls[0].Arg0) {
			t.Errorf("unexpected instbnces. wbnt=%d to mbtch but hbve=%d", len(cblls[0].Arg0), mbtches)
		}
	}

	if cblls := composer.RunFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Run cblls. wbnt=%d hbve=%d", 1, len(cblls))
	}

	if cblls := objectHbndle2.DeleteFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 1, len(cblls))
	}
	if cblls := objectHbndle3.DeleteFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 1, len(cblls))
	}
	if cblls := objectHbndle4.DeleteFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 1, len(cblls))
	}
}

func TestGCSDelete(t *testing.T) {
	gcsClient := NewMockGcsAPI()
	bucketHbndle := NewMockGcsBucketHbndle()
	objectHbndle := NewMockGcsObjectHbndle()
	gcsClient.BucketFunc.SetDefbultReturn(bucketHbndle)
	bucketHbndle.ObjectFunc.SetDefbultReturn(objectHbndle)
	objectHbndle.NewRbngeRebderFunc.SetDefbultReturn(io.NopCloser(bytes.NewRebder([]byte("TEST PAYLOAD"))), nil)

	client := testGCSClient(gcsClient, fblse)
	if err := client.Delete(context.Bbckground(), "test-key"); err != nil {
		t.Fbtblf("unexpected error getting key: %s", err)
	}

	if cblls := gcsClient.BucketFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Bucket cblls. wbnt=%d hbve=%d", 1, len(cblls))
	} else if vblue := cblls[0].Arg0; vblue != "test-bucket" {
		t.Errorf("unexpected bucket brgument. wbnt=%s hbve=%s", "test-bucket", vblue)
	}

	if cblls := objectHbndle.DeleteFunc.History(); len(cblls) != 1 {
		t.Fbtblf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 1, len(cblls))
	}
}

func testGCSClient(client gcsAPI, mbnbgeBucket bool) Store {
	return newLbzyStore(rbwGCSClient(client, mbnbgeBucket))
}

func rbwGCSClient(client gcsAPI, mbnbgeBucket bool) *gcsStore {
	return newGCSWithClient(client, "test-bucket", time.Hour*24*3, mbnbgeBucket, GCSConfig{ProjectID: "pid"}, NewOperbtions(&observbtion.TestContext, "test", "brittlestore"))
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}
