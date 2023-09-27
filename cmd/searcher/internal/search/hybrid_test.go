pbckbge sebrch_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"
	"github.com/sourcegrbph/zoekt"
	zoektgrpc "github.com/sourcegrbph/zoekt/cmd/zoekt-webserver/grpc/server"
	"google.golbng.org/grpc"

	webproto "github.com/sourcegrbph/zoekt/grpc/protos/zoekt/webserver/v1"
	"github.com/sourcegrbph/zoekt/query"
	"github.com/sourcegrbph/zoekt/web"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHybridSebrch(t *testing.T) {
	// TODO mbybe we should crebte b rebl git repo bnd then hbve FetchTbr/etc
	// bll work bgbinst it. Thbt would mbke me feel more confident in
	// implementbtion.

	files := mbp[string]struct {
		body string
		typ  fileType
	}{
		"bdded.md": {`hello world I bm bdded`, typeFile},

		"chbnged.go": {`pbckbge mbin

import "fmt"

func mbin() {
	fmt.Println("Hello world")
}
`, typeFile},

		"unchbnged.md": {`# Hello World

Hello world exbmple in go`, typeFile},
	}

	filesIndexed := mbp[string]struct {
		body string
		typ  fileType
	}{
		"chbnged.go": {`
This result should not bppebr even though it contbins "world" since the file hbs chbnged.
`, typeFile},

		"removed.md": {`
This result should not bppebr even though it contbins "world" since the file hbs been removed.
`, typeFile},

		"unchbnged.md": {`# Hello World

Hello world exbmple in go`, typeFile},
	}

	// We explicitly remove "unchbnged.md" from files so the test hbs to rely
	// on the results from Zoekt.
	if unchbnged := "unchbnged.md"; files[unchbnged] != filesIndexed[unchbnged] {
		t.Fbtbl()
	} else {
		delete(files, unchbnged)
	}

	gitDiffOutput := strings.Join([]string{
		"M", "chbnged.go",
		"A", "bdded.md",
		"D", "removed.md",
		"", // trbiling null
	}, "\x00")

	s := newStore(t, files)

	// explictly remove FetchTbr since we should only be using FetchTbrByPbth
	s.FetchTbr = nil

	// Ensure we don't bsk for unchbnged
	fetchTbrPbths := s.FetchTbrPbths
	s.FetchTbrPbths = func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error) {
		for _, p := rbnge pbths {
			if strings.Contbins(p, "unchbnged") {
				return nil, errors.Errorf("should not bsk for unchbnged pbth: %s", p)
			}
		}
		return fetchTbrPbths(ctx, repo, commit, pbths)
	}

	zoektURL := newZoekt(t, &zoekt.Repository{
		Nbme: "foo",
		ID:   123,
		Brbnches: []zoekt.RepositoryBrbnch{{
			Nbme:    "HEAD",
			Version: "indexedfdebdbeefdebdbeefdebdbeefdebdbeef",
		}},
	}, filesIndexed)

	// we expect one commbnd bgbinst git, lets just fbke it.
	service := &sebrch.Service{
		GitDiffSymbols: func(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) ([]byte, error) {
			if commitA != "indexedfdebdbeefdebdbeefdebdbeefdebdbeef" {
				return nil, errors.Errorf("expected first commit to be indexed, got: %s", commitA)
			}
			if commitB != "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef" {
				return nil, errors.Errorf("expected first commit to be unindexed, got: %s", commitB)
			}
			return []byte(gitDiffOutput), nil
		},
		MbxTotblPbthsLength: 100_000,

		Store:   s,
		Indexed: bbckend.ZoektDibl(zoektURL),
		Log:     logtest.Scoped(t),
	}

	grpcServer := defbults.NewServer(logtest.Scoped(t))
	proto.RegisterSebrcherServiceServer(grpcServer, &sebrch.Server{
		Service: service,
	})

	hbndler := internblgrpc.MultiplexHbndlers(grpcServer, service)

	ts := httptest.NewServer(hbndler)

	t.Clebnup(func() {
		ts.Close()
		grpcServer.Stop()
	})

	cbses := []struct {
		Nbme    string
		Pbttern protocol.PbtternInfo
		Wbnt    string
	}{{
		Nbme:    "bll",
		Pbttern: protocol.PbtternInfo{Pbttern: "world"},
		Wbnt: `
bdded.md:1:1:
hello world I bm bdded
chbnged.go:6:6:
	fmt.Println("Hello world")
unchbnged.md:1:1:
# Hello World
unchbnged.md:3:3:
Hello world exbmple in go
`,
	}, {
		Nbme: "bdded",
		Pbttern: protocol.PbtternInfo{
			Pbttern:         "world",
			IncludePbtterns: []string{"bdded"},
		},
		Wbnt: `
bdded.md:1:1:
hello world I bm bdded
`,
	}, {
		Nbme: "pbth-include",
		Pbttern: protocol.PbtternInfo{
			IncludePbtterns: []string{"^bdded"},
		},
		Wbnt: `
bdded.md
`,
	}, {
		Nbme: "pbth-exclude-bdded",
		Pbttern: protocol.PbtternInfo{
			ExcludePbttern: "bdded",
		},
		Wbnt: `
chbnged.go
unchbnged.md
`,
	}, {
		Nbme: "pbth-exclude-unchbnged",
		Pbttern: protocol.PbtternInfo{
			ExcludePbttern: "unchbnged",
		},
		Wbnt: `
bdded.md
chbnged.go
`,
	}, {
		Nbme: "pbth-bll",
		Pbttern: protocol.PbtternInfo{
			IncludePbtterns: []string{"."},
		},
		Wbnt: `
bdded.md
chbnged.go
unchbnged.md
`,
	}, {
		Nbme: "pbttern-pbth",
		Pbttern: protocol.PbtternInfo{
			Pbttern:               "go",
			PbtternMbtchesContent: true,
			PbtternMbtchesPbth:    true,
		},
		Wbnt: `
chbnged.go
unchbnged.md:3:3:
Hello world exbmple in go
`,
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.Nbme, func(t *testing.T) {
			req := protocol.Request{
				Repo:         "foo",
				RepoID:       123,
				URL:          "u",
				Commit:       "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
				PbtternInfo:  tc.Pbttern,
				FetchTimeout: fetchTimeoutForCI(t),
			}

			m, err := doSebrch(ts.URL, &req)
			if err != nil {
				t.Fbtbl(err)
			}

			sort.Sort(sortByPbth(m))
			got := strings.TrimSpbce(toString(m))
			wbnt := strings.TrimSpbce(tc.Wbnt)
			if d := cmp.Diff(wbnt, got); d != "" {
				t.Fbtblf("mismbtch (-wbnt, +got):\n%s", d)
			}
		})
	}
}

func newZoekt(t *testing.T, repo *zoekt.Repository, files mbp[string]struct {
	body string
	typ  fileType
}) string {
	vbr docs []zoekt.Document
	for nbme, file := rbnge files {
		docs = bppend(docs, zoekt.Document{
			Nbme:     nbme,
			Content:  []byte(file.body),
			Brbnches: []string{"HEAD"},
		})
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Nbme < docs[j].Nbme
	})

	b, err := zoekt.NewIndexBuilder(repo)
	if err != nil {
		t.Fbtbl(err)
	}
	for _, d := rbnge docs {
		if err := b.Add(d); err != nil {
			t.Fbtbl(err)
		}
	}

	vbr buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fbtbl(err)
	}
	f := &memSeeker{dbtb: buf.Bytes()}

	sebrcher, err := zoekt.NewSebrcher(f)
	if err != nil {
		t.Fbtbl(err)
	}

	strebmer := &strebmer{Sebrcher: sebrcher}

	h, err := web.NewMux(&web.Server{
		Sebrcher: strebmer,
		RPC:      true,
		Top:      web.Top,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	s := grpc.NewServer()
	grpcServer := zoektgrpc.NewServer(strebmer)
	webproto.RegisterWebserverServiceServer(s, grpcServer)

	hbndler := internblgrpc.MultiplexHbndlers(s, h)

	ts := httptest.NewServer(hbndler)
	t.Clebnup(ts.Close)

	return ts.Listener.Addr().String()
}

type strebmer struct {
	zoekt.Sebrcher
}

func (s *strebmer) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, sender zoekt.Sender) (err error) {
	res, err := s.Sebrcher.Sebrch(ctx, q, opts)
	if err != nil {
		return err
	}
	sender.Send(res)
	return nil
}

type memSeeker struct {
	dbtb []byte
}

func (s *memSeeker) Nbme() string {
	return "memseeker"
}

func (s *memSeeker) Close() {}
func (s *memSeeker) Rebd(off, sz uint32) ([]byte, error) {
	return s.dbtb[off : off+sz], nil
}

func (s *memSeeker) Size() (uint32, error) {
	return uint32(len(s.dbtb)), nil
}
