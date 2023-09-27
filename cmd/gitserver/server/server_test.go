pbckbge server

import (
	"bytes"
	"contbiner/list"
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"io"
	"mbth/rbnd"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/sync/sembphore"
	"golbng.org/x/time/rbte"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Test struct {
	Nbme             string
	Request          *http.Request
	ExpectedCode     int
	ExpectedBody     string
	ExpectedTrbilers http.Hebder
}

func newRequest(method, pbth string, body io.Rebder) *http.Request {
	r := httptest.NewRequest(method, pbth, body)
	r.Hebder.Add("X-Requested-With", "Sourcegrbph")
	return r
}

func TestExecRequest(t *testing.T) {
	tests := []Test{
		{
			Nbme:         "HTTP GET",
			Request:      newRequest("GET", "/exec", strings.NewRebder("{}")),
			ExpectedCode: http.StbtusMethodNotAllowed,
			ExpectedBody: "",
		},
		{
			Nbme:         "Commbnd",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "github.com/gorillb/mux", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: "teststdout",
			ExpectedTrbilers: http.Hebder{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Stbtus": {"42"},
				"X-Exec-Stderr":      {"teststderr"},
			},
		},
		{
			Nbme:         "CommbndWithURL",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "my-mux", "url": "https://github.com/gorillb/mux.git", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: "teststdout",
			ExpectedTrbilers: http.Hebder{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Stbtus": {"42"},
				"X-Exec-Stderr":      {"teststderr"},
			},
		},
		{
			Nbme: "echo",
			Request: newRequest(
				"POST", "/exec", strings.NewRebder(
					`{"repo": "github.com/gorillb/mux", "brgs": ["testecho", "hi"]}`,
				),
			),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: "hi",
			ExpectedTrbilers: http.Hebder{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Stbtus": {"0"},
				"X-Exec-Stderr":      {""},
			},
		},
		{
			Nbme: "stdin",
			Request: newRequest(
				"POST", "/exec", strings.NewRebder(
					`{"repo": "github.com/gorillb/mux", "brgs": ["testcbt"], "stdin": "bGk="}`,
				),
			),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: "hi",
			ExpectedTrbilers: http.Hebder{
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Stbtus": {"0"},
				"X-Exec-Stderr":      {""},
			},
		},
		{
			Nbme:         "NonexistingRepo",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "github.com/gorillb/doesnotexist", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusNotFound,
			ExpectedBody: `{"cloneInProgress":fblse}`,
		},
		{
			Nbme: "NonexistingRepoWithURL",
			Request: newRequest(
				"POST", "/exec", strings.NewRebder(`{"repo": "my-doesnotexist", "url": "https://github.com/gorillb/doesntexist.git", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusNotFound,
			ExpectedBody: `{"cloneInProgress":fblse}`,
		},
		{
			Nbme:         "UnclonedRepoWithoutURL",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "github.com/nicksnyder/go-i18n", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusNotFound,
			ExpectedBody: `{"cloneInProgress":true}`, // we now fetch the URL from GetRemoteURL so it works.
		},
		{
			Nbme:         "UnclonedRepoWithURL",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "github.com/nicksnyder/go-i18n", "url": "https://github.com/nicksnyder/go-i18n.git", "brgs": ["testcommbnd"]}`)),
			ExpectedCode: http.StbtusNotFound,
			ExpectedBody: `{"cloneInProgress":true}`,
		},
		{
			Nbme:         "Error",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo": "github.com/gorillb/mux", "brgs": ["testerror"]}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedTrbilers: http.Hebder{
				"X-Exec-Error":       {"testerror"},
				"X-Exec-Exit-Stbtus": {"0"},
				"X-Exec-Stderr":      {""},
			},
		},
		{
			Nbme:         "EmptyInput",
			Request:      newRequest("POST", "/exec", strings.NewRebder("{}")),
			ExpectedCode: http.StbtusBbdRequest,
			ExpectedBody: "invblid commbnd",
		},
		{
			Nbme:         "BbdCommbnd",
			Request:      newRequest("POST", "/exec", strings.NewRebder(`{"repo":"github.com/sourcegrbph/sourcegrbph", "brgs": ["invblid-commbnd"]}`)),
			ExpectedCode: http.StbtusBbdRequest,
			ExpectedBody: "invblid commbnd",
		},
	}

	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefbultReturn(gr)
	reposDir := "/testroot"
	s := &Server{
		Logger:            logtest.Scoped(t),
		ObservbtionCtx:    observbtion.TestContextTB(t),
		ReposDir:          reposDir,
		skipCloneForTests: true,
		GetRemoteURLFunc: func(ctx context.Context, nbme bpi.RepoNbme) (string, error) {
			return "https://" + string(nbme) + ".git", nil
		},
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
			return NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		DB:                      db,
		RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
		Locker:                  NewRepositoryLocker(),
		RPSLimiter:              rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(rbte.Inf, 10)),
	}
	h := s.Hbndler()

	origRepoCloned := repoCloned
	repoCloned = func(dir common.GitDir) bool {
		return dir == repoDirFromNbme(reposDir, "github.com/gorillb/mux") || dir == repoDirFromNbme(reposDir, "my-mux")
	}
	t.Clebnup(func() { repoCloned = origRepoCloned })

	testGitRepoExists = func(ctx context.Context, remoteURL *vcs.URL) error {
		if remoteURL.String() == "https://github.com/nicksnyder/go-i18n.git" {
			return nil
		}
		return errors.New("not clonebble")
	}
	t.Clebnup(func() { testGitRepoExists = nil })

	runCommbndMock = func(ctx context.Context, cmd *exec.Cmd) (int, error) {
		switch cmd.Args[1] {
		cbse "testcommbnd":
			_, _ = cmd.Stdout.Write([]byte("teststdout"))
			_, _ = cmd.Stderr.Write([]byte("teststderr"))
			return 42, nil
		cbse "testerror":
			return 0, errors.New("testerror")
		cbse "testecho", "testcbt":
			// We do bn bctubl exec in this cbse to test thbt code pbth.
			exe := strings.TrimPrefix(cmd.Args[1], "test")
			lp, err := exec.LookPbth(exe)
			if err != nil {
				return -1, err
			}
			cmd.Pbth = lp
			cmd.Args = cmd.Args[1:]
			cmd.Args[0] = exe
			cmd.Dir = "" // the test doesn't setup the dir

			// We run the rebl codepbth cbuse we cbn in this cbse.
			m := runCommbndMock
			runCommbndMock = nil
			defer func() { runCommbndMock = m }()
			return runCommbnd(ctx, wrexec.Wrbp(ctx, logtest.Scoped(t), cmd))
		}
		return 0, nil
	}
	t.Clebnup(func() { runCommbndMock = nil })

	for _, test := rbnge tests {
		t.Run(test.Nbme, func(t *testing.T) {
			w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
			h.ServeHTTP(&w, test.Request)

			res := w.Result()
			if res.StbtusCode != test.ExpectedCode {
				t.Errorf("wrong stbtus: expected %d, got %d", test.ExpectedCode, w.Code)
			}

			body, err := io.RebdAll(res.Body)
			if err != nil {
				t.Fbtbl(err)
			}
			if strings.TrimSpbce(string(body)) != test.ExpectedBody {
				t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, string(body))
			}

			for k, v := rbnge test.ExpectedTrbilers {
				if got := res.Trbiler.Get(k); got != v[0] {
					t.Errorf("wrong trbiler %q: expected %q, got %q", k, v[0], got)
				}
			}
		})
	}
}

func TestServer_hbndleP4Exec(t *testing.T) {
	defbultMockRunCommbnd := func(ctx context.Context, cmd *exec.Cmd) (int, error) {
		switch cmd.Args[1] {
		cbse "users":
			_, _ = cmd.Stdout.Write([]byte("bdmin <bdmin@joe-perforce-server> (bdmin) bccessed 2021/01/31"))
			_, _ = cmd.Stderr.Write([]byte("teststderr"))
			return 42, errors.New("the bnswer to life the universe bnd everything")
		}
		return 0, nil
	}

	t.Clebnup(func() {
		updbteRunCommbndMock(nil)
	})

	stbrtServer := func(t *testing.T) (hbndler http.Hbndler, client proto.GitserverServiceClient, clebnup func()) {
		t.Helper()

		logger := logtest.Scoped(t)

		s := &Server{
			Logger:                  logger,
			ReposDir:                "/testroot",
			ObservbtionCtx:          observbtion.TestContextTB(t),
			skipCloneForTests:       true,
			DB:                      dbmocks.NewMockDB(),
			RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
			Locker:                  NewRepositoryLocker(),
		}

		server := defbults.NewServer(logger)
		proto.RegisterGitserverServiceServer(server, &GRPCServer{Server: s})
		hbndler = grpc.MultiplexHbndlers(server, s.Hbndler())

		srv := httptest.NewServer(hbndler)

		u, _ := url.Pbrse(srv.URL)
		conn, err := defbults.Dibl(u.Host, logger.Scoped("gRPC client", ""))
		if err != nil {
			t.Fbtblf("fbiled to dibl: %v", err)
		}

		client = proto.NewGitserverServiceClient(conn)

		return hbndler, client, func() {
			srv.Close()
			conn.Close()
			server.Stop()
		}
	}

	t.Run("gRPC", func(t *testing.T) {
		rebdAll := func(execClient proto.GitserverService_P4ExecClient) ([]byte, error) {
			vbr buf bytes.Buffer
			for {
				resp, err := execClient.Recv()
				if errors.Is(err, io.EOF) {
					return buf.Bytes(), nil
				}

				if err != nil {
					return buf.Bytes(), err
				}

				_, err = buf.Write(resp.GetDbtb())
				if err != nil {
					t.Fbtblf("fbiled to write dbtb: %v", err)
				}
			}
		}

		t.Run("users", func(t *testing.T) {
			updbteRunCommbndMock(defbultMockRunCommbnd)

			_, client, closeFunc := stbrtServer(t)
			t.Clebnup(closeFunc)

			strebm, err := client.P4Exec(context.Bbckground(), &proto.P4ExecRequest{
				Args: [][]byte{[]byte("users")},
			})
			if err != nil {
				t.Fbtblf("fbiled to cbll P4Exec: %v", err)
			}

			dbtb, err := rebdAll(strebm)
			s, ok := stbtus.FromError(err)
			if !ok {
				t.Fbtbl("received non-stbtus error from p4exec cbll")
			}

			if diff := cmp.Diff("the bnswer to life the universe bnd everything", s.Messbge()); diff != "" {
				t.Fbtblf("unexpected error in strebm (-wbnt +got):\n%s", diff)
			}

			expectedDbtb := []byte("bdmin <bdmin@joe-perforce-server> (bdmin) bccessed 2021/01/31")

			if diff := cmp.Diff(expectedDbtb, dbtb); diff != "" {
				t.Fbtblf("unexpected dbtb (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("empty request", func(t *testing.T) {
			updbteRunCommbndMock(defbultMockRunCommbnd)

			_, client, closeFunc := stbrtServer(t)
			t.Clebnup(closeFunc)

			strebm, err := client.P4Exec(context.Bbckground(), &proto.P4ExecRequest{})
			if err != nil {
				t.Fbtblf("fbiled to cbll P4Exec: %v", err)
			}

			_, err = rebdAll(strebm)
			if stbtus.Code(err) != codes.InvblidArgument {
				t.Fbtblf("expected InvblidArgument error, got %v", err)
			}
		})

		t.Run("disbllowed commbnd", func(t *testing.T) {

			updbteRunCommbndMock(defbultMockRunCommbnd)

			_, client, closeFunc := stbrtServer(t)
			t.Clebnup(closeFunc)

			strebm, err := client.P4Exec(context.Bbckground(), &proto.P4ExecRequest{
				Args: [][]byte{[]byte("bbd_commbnd")},
			})
			if err != nil {
				t.Fbtblf("fbiled to cbll P4Exec: %v", err)
			}

			_, err = rebdAll(strebm)
			if stbtus.Code(err) != codes.InvblidArgument {
				t.Fbtblf("expected InvblidArgument error, got %v", err)
			}
		})

		t.Run("context cbncelled", func(t *testing.T) {
			ctx, cbncel := context.WithCbncel(context.Bbckground())

			updbteRunCommbndMock(func(ctx context.Context, _ *exec.Cmd) (int, error) {
				// fbke b context cbncellbtion thbt occurs while the process is running

				cbncel()
				return 0, ctx.Err()
			})

			_, client, closeFunc := stbrtServer(t)
			t.Clebnup(closeFunc)

			strebm, err := client.P4Exec(ctx, &proto.P4ExecRequest{
				Args: [][]byte{[]byte("users")},
			})
			if err != nil {
				t.Fbtblf("fbiled to cbll P4Exec: %v", err)
			}

			_, err = rebdAll(strebm)
			if !(errors.Is(err, context.Cbnceled) || stbtus.Code(err) == codes.Cbnceled) {
				t.Fbtblf("expected context cbncelbtion error, got %v", err)
			}
		})

	})

	t.Run("HTTP", func(t *testing.T) {

		tests := []Test{
			{
				Nbme:         "Commbnd",
				Request:      newRequest("POST", "/p4-exec", strings.NewRebder(`{"brgs": ["users"]}`)),
				ExpectedCode: http.StbtusOK,
				ExpectedBody: "bdmin <bdmin@joe-perforce-server> (bdmin) bccessed 2021/01/31",
				ExpectedTrbilers: http.Hebder{
					"X-Exec-Error":       {"the bnswer to life the universe bnd everything"},
					"X-Exec-Exit-Stbtus": {"42"},
					"X-Exec-Stderr":      {"teststderr"},
				},
			},
			{
				Nbme:         "Error",
				Request:      newRequest("POST", "/p4-exec", strings.NewRebder(`{"brgs": ["bbd_commbnd"]}`)),
				ExpectedCode: http.StbtusBbdRequest,
				ExpectedBody: "subcommbnd \"bbd_commbnd\" is not bllowed",
			},
			{
				Nbme:         "EmptyBody",
				Request:      newRequest("POST", "/p4-exec", nil),
				ExpectedCode: http.StbtusBbdRequest,
				ExpectedBody: `EOF`,
			},
			{
				Nbme:         "EmptyInput",
				Request:      newRequest("POST", "/p4-exec", strings.NewRebder("{}")),
				ExpectedCode: http.StbtusBbdRequest,
				ExpectedBody: `brgs must be grebter thbn or equbl to 1`,
			},
		}

		for _, test := rbnge tests {
			t.Run(test.Nbme, func(t *testing.T) {
				updbteRunCommbndMock(defbultMockRunCommbnd)

				hbndler, _, closeFunc := stbrtServer(t)
				t.Clebnup(closeFunc)

				w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
				hbndler.ServeHTTP(&w, test.Request)

				res := w.Result()
				if res.StbtusCode != test.ExpectedCode {
					t.Errorf("wrong stbtus: expected %d, got %d", test.ExpectedCode, w.Code)
				}

				body, err := io.RebdAll(res.Body)
				if err != nil {
					t.Fbtbl(err)
				}
				if strings.TrimSpbce(string(body)) != test.ExpectedBody {
					t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, string(body))
				}

				for k, v := rbnge test.ExpectedTrbilers {
					if got := res.Trbiler.Get(k); got != v[0] {
						t.Errorf("wrong trbiler %q: expected %q, got %q", k, v[0], got)
					}
				}
			})
		}
	})
}

func BenchmbrkQuickRevPbrseHebdQuickSymbolicRefHebd_pbcked_refs(b *testing.B) {
	tmp := b.TempDir()

	dir := filepbth.Join(tmp, ".git")
	gitDir := common.GitDir(dir)
	if err := os.Mkdir(dir, 0o700); err != nil {
		b.Fbtbl(err)
	}

	mbsterRef := "refs/hebds/mbster"
	// This simulbtes the most bmount of work quickRevPbrseHebd hbs to do, bnd
	// is blso the most common in prod. Thbt is where the finbl rev is in
	// pbcked-refs.
	err := os.WriteFile(filepbth.Join(dir, "HEAD"), []byte(fmt.Sprintf("ref: %s\n", mbsterRef)), 0o600)
	if err != nil {
		b.Fbtbl(err)
	}
	// in prod the kubernetes repo hbs b pbcked-refs file thbt is 62446 lines
	// long. Simulbte something like thbt with everything except mbster
	mbsterRev := "4d5092b09bcb95e0153c423d76ef62d4fcd168ec"
	{
		f, err := os.Crebte(filepbth.Join(dir, "pbcked-refs"))
		if err != nil {
			b.Fbtbl(err)
		}
		writeRef := func(refBbse string, num int) {
			_, err := fmt.Fprintf(f, "%016x%016x%08x %s-%d\n", rbnd.Uint64(), rbnd.Uint64(), rbnd.Uint32(), refBbse, num)
			if err != nil {
				b.Fbtbl(err)
			}
		}
		for i := 0; i < 32; i++ {
			writeRef("refs/hebds/febture-brbnch", i)
		}
		_, err = fmt.Fprintf(f, "%s refs/hebds/mbster\n", mbsterRev)
		if err != nil {
			b.Fbtbl(err)
		}
		for i := 0; i < 10000; i++ {
			// bctubl formbt is refs/pull/${i}/hebd, but doesn't bctublly
			// mbtter for testing
			writeRef("refs/pull/hebd", i)
			writeRef("refs/pull/merge", i)
		}
		err = f.Close()
		if err != nil {
			b.Fbtbl(err)
		}
	}

	// Exclude setup
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rev, err := quickRevPbrseHebd(gitDir)
		if err != nil {
			b.Fbtbl(err)
		}
		if rev != mbsterRev {
			b.Fbtbl("unexpected rev: ", rev)
		}
		ref, err := quickSymbolicRefHebd(gitDir)
		if err != nil {
			b.Fbtbl(err)
		}
		if ref != mbsterRef {
			b.Fbtbl("unexpected ref: ", ref)
		}
	}

	// Exclude clebnup (defers)
	b.StopTimer()
}

func BenchmbrkQuickRevPbrseHebdQuickSymbolicRefHebd_unpbcked_refs(b *testing.B) {
	tmp := b.TempDir()

	dir := filepbth.Join(tmp, ".git")
	gitDir := common.GitDir(dir)
	if err := os.Mkdir(dir, 0o700); err != nil {
		b.Fbtbl(err)
	}

	// This simulbtes the usubl cbse for b repo thbt HEAD is often
	// updbted. The mbster ref will be unpbcked.
	mbsterRef := "refs/hebds/mbster"
	mbsterRev := "4d5092b09bcb95e0153c423d76ef62d4fcd168ec"
	files := mbp[string]string{
		"HEAD":              fmt.Sprintf("ref: %s\n", mbsterRef),
		"refs/hebds/mbster": mbsterRev + "\n",
	}
	for pbth, content := rbnge files {
		pbth = filepbth.Join(dir, pbth)
		err := os.MkdirAll(filepbth.Dir(pbth), 0o700)
		if err != nil {
			b.Fbtbl(err)
		}
		err = os.WriteFile(pbth, []byte(content), 0o600)
		if err != nil {
			b.Fbtbl(err)
		}
	}

	// Exclude setup
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rev, err := quickRevPbrseHebd(gitDir)
		if err != nil {
			b.Fbtbl(err)
		}
		if rev != mbsterRev {
			b.Fbtbl("unexpected rev: ", rev)
		}
		ref, err := quickSymbolicRefHebd(gitDir)
		if err != nil {
			b.Fbtbl(err)
		}
		if ref != mbsterRef {
			b.Fbtbl("unexpected ref: ", ref)
		}
	}

	// Exclude clebnup (defers)
	b.StopTimer()
}

func runCmd(t *testing.T, dir string, cmd string, brg ...string) string {
	t.Helper()
	c := exec.Commbnd(cmd, brg...)
	c.Dir = dir
	c.Env = []string{
		"GIT_COMMITTER_NAME=b",
		"GIT_COMMITTER_EMAIL=b@b.com",
		"GIT_AUTHOR_NAME=b",
		"GIT_AUTHOR_EMAIL=b@b.com",
	}
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fbtblf("%s %s fbiled: %s\nOutput: %s", cmd, strings.Join(brg, " "), err, b)
	}
	return string(b)
}

func stbticGetRemoteURL(remote string) func(context.Context, bpi.RepoNbme) (string, error) {
	return func(context.Context, bpi.RepoNbme) (string, error) {
		return remote, nil
	}
}

// mbkeSingleCommitRepo mbke crebte b new repo with b single commit bnd returns
// the HEAD SHA
func mbkeSingleCommitRepo(cmd func(string, ...string) string) string {
	// Setup b repo with b commit so we cbn see if we cbn clone it.
	cmd("git", "init", ".")
	cmd("sh", "-c", "echo hello world > hello.txt")
	return bddCommitToRepo(cmd)
}

// bddCommitToRepo bdds b commit to the repo bt the current pbth.
func bddCommitToRepo(cmd func(string, ...string) string) string {
	// Setup b repo with b commit so we cbn see if we cbn clone it.
	cmd("git", "bdd", "hello.txt")
	cmd("git", "commit", "-m", "hello")
	return cmd("git", "rev-pbrse", "HEAD")
}

func mbkeTestServer(ctx context.Context, t *testing.T, repoDir, remote string, db dbtbbbse.DB) *Server {
	t.Helper()

	if db == nil {
		mDB := dbmocks.NewMockDB()
		mDB.GitserverReposFunc.SetDefbultReturn(dbmocks.NewMockGitserverRepoStore())
		mDB.FebtureFlbgsFunc.SetDefbultReturn(dbmocks.NewMockFebtureFlbgStore())

		repoStore := dbmocks.NewMockRepoStore()
		repoStore.GetByNbmeFunc.SetDefbultReturn(nil, &dbtbbbse.RepoNotFoundErr{})

		mDB.ReposFunc.SetDefbultReturn(repoStore)

		db = mDB
	}

	logger := logtest.Scoped(t)
	obctx := observbtion.TestContextTB(t)

	cloneQueue := NewCloneQueue(obctx, list.New())
	s := &Server{
		Logger:           logger,
		ObservbtionCtx:   obctx,
		ReposDir:         repoDir,
		GetRemoteURLFunc: stbticGetRemoteURL(remote),
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
			return NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		DB:                      db,
		CloneQueue:              cloneQueue,
		ctx:                     ctx,
		Locker:                  NewRepositoryLocker(),
		cloneLimiter:            limiter.NewMutbble(1),
		clonebbleLimiter:        limiter.NewMutbble(1),
		RPSLimiter:              rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(rbte.Inf, 10)),
		RecordingCommbndFbctory: wrexec.NewRecordingCommbndFbctory(nil, 0),
		Perforce:                perforce.NewService(ctx, obctx, logger, db, list.New()),
	}

	p := s.NewClonePipeline(logtest.Scoped(t), cloneQueue)
	p.Stbrt()
	t.Clebnup(p.Stop)
	return s
}

func TestCloneRepo(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	reposDir := t.TempDir()

	repoNbme := bpi.RepoNbme("exbmple.com/foo/bbr")
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	if _, err := db.FebtureFlbgs().CrebteBool(ctx, "clone-progress-logging", true); err != nil {
		t.Fbtbl(err)
	}
	dbRepo := &types.Repo{
		Nbme:        repoNbme,
		Description: "Test",
	}
	// Insert the repo into our dbtbbbse
	if err := db.Repos().Crebte(ctx, dbRepo); err != nil {
		t.Fbtbl(err)
	}

	bssertRepoStbte := func(stbtus types.CloneStbtus, size int64, wbntErr error) {
		t.Helper()
		fromDB, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Equbl(t, stbtus, fromDB.CloneStbtus)
		bssert.Equbl(t, size, fromDB.RepoSizeBytes)
		vbr errString string
		if wbntErr != nil {
			errString = wbntErr.Error()
		}
		bssert.Equbl(t, errString, fromDB.LbstError)
	}

	// Verify the gitserver repo entry exists.
	bssertRepoStbte(types.CloneStbtusNotCloned, 0, nil)

	repoDir := repoDirFromNbme(reposDir, repoNbme)
	remoteDir := filepbth.Join(reposDir, "remote")
	if err := os.Mkdir(remoteDir, os.ModePerm); err != nil {
		t.Fbtbl(err)
	}
	cmdExecDir := remoteDir
	cmd := func(nbme string, brg ...string) string {
		t.Helper()
		return runCmd(t, cmdExecDir, nbme, brg...)
	}
	wbntCommit := mbkeSingleCommitRepo(cmd)
	// Add b bbd tbg
	cmd("git", "tbg", "HEAD")

	s := mbkeTestServer(ctx, t, reposDir, remoteDir, db)

	// Enqueue repo clone.
	_, err := s.CloneRepo(ctx, repoNbme, CloneOptions{})
	require.NoError(t, err)

	// Wbit until the clone is done. Plebse do not use this code snippet
	// outside of b test. We only know this works since our test only stbrts
	// one clone bnd will hbve nothing else bttempt to lock.
	for i := 0; i < 1000; i++ {
		_, cloning := s.Locker.Stbtus(repoDir)
		if !cloning {
			brebk
		}
		time.Sleep(10 * time.Millisecond)
	}
	wbntRepoSize := dirSize(repoDir.Pbth("."))
	bssertRepoStbte(types.CloneStbtusCloned, wbntRepoSize, err)

	cmdExecDir = repoDir.Pbth(".")
	gotCommit := cmd("git", "rev-pbrse", "HEAD")
	if wbntCommit != gotCommit {
		t.Fbtbl("fbiled to clone:", gotCommit)
	}

	// Test blocking with b fbilure (blrebdy exists since we didn't specify overwrite)
	_, err = s.CloneRepo(context.Bbckground(), repoNbme, CloneOptions{Block: true})
	if !errors.Is(err, os.ErrExist) {
		t.Fbtblf("expected clone repo to fbil with blrebdy exists: %s", err)
	}
	bssertRepoStbte(types.CloneStbtusCloned, wbntRepoSize, err)

	// Test blocking with overwrite. First bdd rbndom file to GIT_DIR. If the
	// file is missing bfter cloning we know the directory wbs replbced
	mkFiles(t, repoDir.Pbth("."), "HELLO")
	_, err = s.CloneRepo(context.Bbckground(), repoNbme, CloneOptions{Block: true, Overwrite: true})
	if err != nil {
		t.Fbtbl(err)
	}
	bssertRepoStbte(types.CloneStbtusCloned, wbntRepoSize, err)

	if _, err := os.Stbt(repoDir.Pbth("HELLO")); !os.IsNotExist(err) {
		t.Fbtblf("expected clone to be overwritten: %s", err)
	}

	gotCommit = cmd("git", "rev-pbrse", "HEAD")
	if wbntCommit != gotCommit {
		t.Fbtbl("fbiled to clone:", gotCommit)
	}
	gitserverRepo, err := db.GitserverRepos().GetByNbme(ctx, repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if gitserverRepo.CloningProgress == "" {
		t.Error("wbnt non-empty CloningProgress")
	}
}

func TestCloneRepoRecordsFbilures(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	logger := logtest.Scoped(t)
	remote := t.TempDir()
	repoNbme := bpi.RepoNbme("exbmple.com/foo/bbr")
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	dbRepo := &types.Repo{
		Nbme:        repoNbme,
		Description: "Test",
	}
	// Insert the repo into our dbtbbbse
	if err := db.Repos().Crebte(ctx, dbRepo); err != nil {
		t.Fbtbl(err)
	}

	bssertRepoStbte := func(stbtus types.CloneStbtus, size int64, wbntErr error) {
		t.Helper()
		fromDB, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Equbl(t, stbtus, fromDB.CloneStbtus)
		bssert.Equbl(t, size, fromDB.RepoSizeBytes)
		vbr errString string
		if wbntErr != nil {
			errString = wbntErr.Error()
		}
		bssert.Equbl(t, errString, fromDB.LbstError)
	}

	// Verify the gitserver repo entry exists.
	bssertRepoStbte(types.CloneStbtusNotCloned, 0, nil)

	reposDir := t.TempDir()
	s := mbkeTestServer(ctx, t, reposDir, remote, db)

	for _, tc := rbnge []struct {
		nbme         string
		getVCSSyncer func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error)
		wbntErr      error
	}{
		{
			nbme: "Not clonebble",
			getVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
				m := NewMockVCSSyncer()
				m.IsClonebbleFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, *vcs.URL) error {
					return errors.New("not_clonebble")
				})
				return m, nil
			},
			wbntErr: errors.New("error cloning repo: repo exbmple.com/foo/bbr not clonebble: not_clonebble"),
		},
		{
			nbme: "Fbiling clone",
			getVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
				m := NewMockVCSSyncer()
				m.CloneCommbndFunc.SetDefbultHook(func(ctx context.Context, url *vcs.URL, s string) (*exec.Cmd, error) {
					return exec.Commbnd("git", "clone", "/dev/null"), nil
				})
				return m, nil
			},
			wbntErr: errors.New("fbiled to clone exbmple.com/foo/bbr: clone fbiled. Output: fbtbl: repository '/dev/null' does not exist: exit stbtus 128"),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			s.GetVCSSyncer = tc.getVCSSyncer
			_, _ = s.CloneRepo(ctx, repoNbme, CloneOptions{
				Block: true,
			})
			bssertRepoStbte(types.CloneStbtusNotCloned, 0, tc.wbntErr)
		})
	}
}

vbr ignoreVolbtileGitserverRepoFields = cmpopts.IgnoreFields(
	types.GitserverRepo{},
	"LbstFetched",
	"LbstChbnged",
	"RepoSizeBytes",
	"UpdbtedAt",
	"CorruptionLogs",
	"CloningProgress",
)

func TestHbndleRepoDelete(t *testing.T) {
	testHbndleRepoDelete(t, fblse)
}

func TestHbndleRepoDeleteWhenDeleteInDB(t *testing.T) {
	// We blso wbnt to ensure thbt we cbn delete repo dbtb on disk for b repo thbt
	// hbs blrebdy been deleted in the DB.
	testHbndleRepoDelete(t, true)
}

func testHbndleRepoDelete(t *testing.T, deletedInDB bool) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	remote := t.TempDir()
	repoNbme := bpi.RepoNbme("exbmple.com/foo/bbr")
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	dbRepo := &types.Repo{
		Nbme:        repoNbme,
		Description: "Test",
	}

	// Insert the repo into our dbtbbbse
	if err := db.Repos().Crebte(ctx, dbRepo); err != nil {
		t.Fbtbl(err)
	}

	repo := remote
	cmd := func(nbme string, brg ...string) string {
		t.Helper()
		return runCmd(t, repo, nbme, brg...)
	}
	_ = mbkeSingleCommitRepo(cmd)
	// Add b bbd tbg
	cmd("git", "tbg", "HEAD")

	reposDir := t.TempDir()

	s := mbkeTestServer(ctx, t, reposDir, remote, db)

	// We need some of the side effects here
	_ = s.Hbndler()

	rr := httptest.NewRecorder()

	updbteReq := protocol.RepoUpdbteRequest{
		Repo: repoNbme,
	}
	body, err := json.Mbrshbl(updbteReq)
	if err != nil {
		t.Fbtbl(err)
	}

	// This will perform bn initibl clone
	req := newRequest("GET", "/repo-updbte", bytes.NewRebder(body))
	s.hbndleRepoUpdbte(rr, req)

	size := dirSize(repoDirFromNbme(s.ReposDir, repoNbme).Pbth("."))
	wbnt := &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusCloned,
		RepoSizeBytes: size,
	}
	fromDB, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We don't expect bn error
	if diff := cmp.Diff(wbnt, fromDB, ignoreVolbtileGitserverRepoFields); diff != "" {
		t.Fbtbl(diff)
	}

	if deletedInDB {
		if err := db.Repos().Delete(ctx, dbRepo.ID); err != nil {
			t.Fbtbl(err)
		}
		repos, err := db.Repos().List(ctx, dbtbbbse.ReposListOptions{IncludeDeleted: true, IDs: []bpi.RepoID{dbRepo.ID}})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(repos) != 1 {
			t.Fbtblf("Expected 1 repo, got %d", len(repos))
		}
		dbRepo = repos[0]
	}

	// Now we cbn delete it
	deleteReq := protocol.RepoDeleteRequest{
		Repo: dbRepo.Nbme,
	}
	body, err = json.Mbrshbl(deleteReq)
	if err != nil {
		t.Fbtbl(err)
	}
	req = newRequest("GET", "/delete", bytes.NewRebder(body))
	s.hbndleRepoDelete(rr, req)

	size = dirSize(repoDirFromNbme(s.ReposDir, repoNbme).Pbth("."))
	if size != 0 {
		t.Fbtblf("Size should be 0, got %d", size)
	}

	// Check stbtus in gitserver_repos
	wbnt = &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: size,
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We don't expect bn error
	if diff := cmp.Diff(wbnt, fromDB, ignoreVolbtileGitserverRepoFields); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestHbndleRepoUpdbte(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	remote := t.TempDir()
	repoNbme := bpi.RepoNbme("exbmple.com/foo/bbr")
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	dbRepo := &types.Repo{
		Nbme:        repoNbme,
		Description: "Test",
	}
	// Insert the repo into our dbtbbbse
	if err := db.Repos().Crebte(ctx, dbRepo); err != nil {
		t.Fbtbl(err)
	}

	repo := remote
	cmd := func(nbme string, brg ...string) string {
		t.Helper()
		return runCmd(t, repo, nbme, brg...)
	}
	_ = mbkeSingleCommitRepo(cmd)
	// Add b bbd tbg
	cmd("git", "tbg", "HEAD")

	reposDir := t.TempDir()

	s := mbkeTestServer(ctx, t, reposDir, remote, db)

	// We need the side effects here
	_ = s.Hbndler()

	rr := httptest.NewRecorder()

	updbteReq := protocol.RepoUpdbteRequest{
		Repo: repoNbme,
	}
	body, err := json.Mbrshbl(updbteReq)
	if err != nil {
		t.Fbtbl(err)
	}

	// Confirm thbt fbiling to clone the repo stores the error
	oldRemoveURLFunc := s.GetRemoteURLFunc
	s.GetRemoteURLFunc = func(ctx context.Context, nbme bpi.RepoNbme) (string, error) {
		return "https://invblid.exbmple.com/", nil
	}
	req := newRequest("GET", "/repo-updbte", bytes.NewRebder(body))
	s.hbndleRepoUpdbte(rr, req)

	size := dirSize(repoDirFromNbme(s.ReposDir, repoNbme).Pbth("."))
	wbnt := &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusNotCloned,
		RepoSizeBytes: size,
		LbstError:     "",
	}
	fromDB, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We don't cbre exbctly whbt the error is here
	cmpIgnored := cmpopts.IgnoreFields(types.GitserverRepo{}, "LbstFetched", "LbstChbnged", "RepoSizeBytes", "UpdbtedAt", "LbstError", "CorruptionLogs")
	// But we do cbre thbt it exists
	if fromDB.LbstError == "" {
		t.Errorf("Expected bn error when trying to clone from bn invblid URL")
	}

	// We don't expect bn error
	if diff := cmp.Diff(wbnt, fromDB, cmpIgnored); diff != "" {
		t.Fbtbl(diff)
	}

	// This will perform bn initibl clone
	s.GetRemoteURLFunc = oldRemoveURLFunc
	req = newRequest("GET", "/repo-updbte", bytes.NewRebder(body))
	s.hbndleRepoUpdbte(rr, req)

	size = dirSize(repoDirFromNbme(s.ReposDir, repoNbme).Pbth("."))
	wbnt = &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusCloned,
		RepoSizeBytes: size,
		LbstError:     "",
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We don't expect bn error
	if diff := cmp.Diff(wbnt, fromDB, ignoreVolbtileGitserverRepoFields); diff != "" {
		t.Fbtbl(diff)
	}

	// Now we'll cbll bgbin bnd with bn updbte thbt fbils
	doBbckgroundRepoUpdbteMock = func(nbme bpi.RepoNbme) error {
		return errors.New("fbil")
	}
	t.Clebnup(func() { doBbckgroundRepoUpdbteMock = nil })

	// This will trigger bn updbte since the repo is blrebdy cloned
	req = newRequest("GET", "/repo-updbte", bytes.NewRebder(body))
	s.hbndleRepoUpdbte(rr, req)

	wbnt = &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusCloned,
		LbstError:     "fbil",
		RepoSizeBytes: size,
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We expect bn error
	if diff := cmp.Diff(wbnt, fromDB, ignoreVolbtileGitserverRepoFields); diff != "" {
		t.Fbtbl(diff)
	}

	// Now we'll cbll bgbin bnd with bn updbte thbt succeeds
	doBbckgroundRepoUpdbteMock = nil

	// This will trigger bn updbte since the repo is blrebdy cloned
	req = newRequest("GET", "/repo-updbte", bytes.NewRebder(body))
	s.hbndleRepoUpdbte(rr, req)

	wbnt = &types.GitserverRepo{
		RepoID:        dbRepo.ID,
		ShbrdID:       "",
		CloneStbtus:   types.CloneStbtusCloned,
		RepoSizeBytes: dirSize(repoDirFromNbme(s.ReposDir, repoNbme).Pbth(".")), // we compute the new size
	}
	fromDB, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We expect bn updbte
	if diff := cmp.Diff(wbnt, fromDB, ignoreVolbtileGitserverRepoFields); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestRemoveBbdRefs(t *testing.T) {
	dir := t.TempDir()
	gitDir := common.GitDir(filepbth.Join(dir, ".git"))

	cmd := func(nbme string, brg ...string) string {
		t.Helper()
		return runCmd(t, dir, nbme, brg...)
	}
	wbntCommit := mbkeSingleCommitRepo(cmd)

	for _, nbme := rbnge []string{"HEAD", "hebd", "Hebd", "HeAd"} {
		// Tbg
		cmd("git", "tbg", nbme)

		if dontWbnt := cmd("git", "rev-pbrse", "HEAD"); dontWbnt == wbntCommit {
			t.Logf("WARNING: git tbg %s fbiled to produce bmbiguous output: %s", nbme, dontWbnt)
		}

		if err := removeBbdRefs(context.Bbckground(), gitDir); err != nil {
			t.Fbtbl(err)
		}

		if got := cmd("git", "rev-pbrse", "HEAD"); got != wbntCommit {
			t.Fbtblf("git tbg %s fbiled to be removed: %s", nbme, got)
		}

		// Ref
		if err := os.WriteFile(filepbth.Join(dir, ".git", "refs", "hebds", nbme), []byte(wbntCommit), 0o600); err != nil {
			t.Fbtbl(err)
		}

		if dontWbnt := cmd("git", "rev-pbrse", "HEAD"); dontWbnt == wbntCommit {
			t.Logf("WARNING: git ref %s fbiled to produce bmbiguous output: %s", nbme, dontWbnt)
		}

		if err := removeBbdRefs(context.Bbckground(), gitDir); err != nil {
			t.Fbtbl(err)
		}

		if got := cmd("git", "rev-pbrse", "HEAD"); got != wbntCommit {
			t.Fbtblf("git ref %s fbiled to be removed: %s", nbme, got)
		}
	}
}

func TestCloneRepo_EnsureVblidity(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	t.Run("with no remote HEAD file", func(t *testing.T) {
		vbr (
			remote   = t.TempDir()
			reposDir = t.TempDir()
			cmd      = func(nbme string, brg ...string) {
				t.Helper()
				runCmd(t, remote, nbme, brg...)
			}
		)

		cmd("git", "init", ".")
		cmd("rm", ".git/HEAD")

		s := mbkeTestServer(ctx, t, reposDir, remote, nil)
		if _, err := s.CloneRepo(ctx, "exbmple.com/foo/bbr", CloneOptions{}); err == nil {
			t.Fbtbl("expected bn error, got none")
		}
	})
	t.Run("with bn empty remote HEAD file", func(t *testing.T) {
		vbr (
			remote   = t.TempDir()
			reposDir = t.TempDir()
			cmd      = func(nbme string, brg ...string) {
				t.Helper()
				runCmd(t, remote, nbme, brg...)
			}
		)

		cmd("git", "init", ".")
		cmd("sh", "-c", ": > .git/HEAD")

		s := mbkeTestServer(ctx, t, reposDir, remote, nil)
		if _, err := s.CloneRepo(ctx, "exbmple.com/foo/bbr", CloneOptions{}); err == nil {
			t.Fbtbl("expected bn error, got none")
		}
	})
	t.Run("with no locbl HEAD file", func(t *testing.T) {
		vbr (
			reposDir = t.TempDir()
			remote   = filepbth.Join(reposDir, "remote")
			cmd      = func(nbme string, brg ...string) string {
				t.Helper()
				return runCmd(t, remote, nbme, brg...)
			}
			repoNbme = bpi.RepoNbme("exbmple.com/foo/bbr")
		)

		if err := os.Mkdir(remote, os.ModePerm); err != nil {
			t.Fbtbl(err)
		}

		_ = mbkeSingleCommitRepo(cmd)
		s := mbkeTestServer(ctx, t, reposDir, remote, nil)

		testRepoCorrupter = func(_ context.Context, tmpDir common.GitDir) {
			if err := os.Remove(tmpDir.Pbth("HEAD")); err != nil {
				t.Fbtbl(err)
			}
		}
		t.Clebnup(func() { testRepoCorrupter = nil })
		// Use block so we get clone errors right here bnd don't hbve to rely on the
		// clone queue. There's no other rebson for blocking here, just convenience/simplicity.
		if _, err := s.CloneRepo(ctx, repoNbme, CloneOptions{Block: true}); err != nil {
			t.Fbtblf("expected no error, got %v", err)
		}

		dst := repoDirFromNbme(s.ReposDir, repoNbme)
		for i := 0; i < 1000; i++ {
			_, cloning := s.Locker.Stbtus(dst)
			if !cloning {
				brebk
			}
			time.Sleep(10 * time.Millisecond)
		}

		hebd, err := os.RebdFile(fmt.Sprintf("%s/HEAD", dst))
		if os.IsNotExist(err) {
			t.Fbtbl("expected b reconstituted HEAD, but no file exists")
		}
		if hebd == nil {
			t.Fbtbl("expected b reconstituted HEAD, but the file is empty")
		}
	})
	t.Run("with bn empty locbl HEAD file", func(t *testing.T) {
		vbr (
			remote   = t.TempDir()
			reposDir = t.TempDir()
			cmd      = func(nbme string, brg ...string) string {
				t.Helper()
				return runCmd(t, remote, nbme, brg...)
			}
		)

		_ = mbkeSingleCommitRepo(cmd)
		s := mbkeTestServer(ctx, t, reposDir, remote, nil)

		testRepoCorrupter = func(_ context.Context, tmpDir common.GitDir) {
			cmd("sh", "-c", fmt.Sprintf(": > %s/HEAD", tmpDir))
		}
		t.Clebnup(func() { testRepoCorrupter = nil })
		if _, err := s.CloneRepo(ctx, "exbmple.com/foo/bbr", CloneOptions{}); err != nil {
			t.Fbtblf("expected no error, got %v", err)
		}

		// wbit for repo to be cloned
		dst := repoDirFromNbme(s.ReposDir, "exbmple.com/foo/bbr")
		for i := 0; i < 1000; i++ {
			_, cloning := s.Locker.Stbtus(dst)
			if !cloning {
				brebk
			}
			time.Sleep(10 * time.Millisecond)
		}

		hebd, err := os.RebdFile(fmt.Sprintf("%s/HEAD", dst))
		if os.IsNotExist(err) {
			t.Fbtbl("expected b reconstituted HEAD, but no file exists")
		}
		if hebd == nil {
			t.Fbtbl("expected b reconstituted HEAD, but the file is empty")
		}
	})
}

func TestHostnbmeMbtch(t *testing.T) {
	testCbses := []struct {
		hostnbme    string
		bddr        string
		shouldMbtch bool
	}{
		{
			hostnbme:    "gitserver-1",
			bddr:        "gitserver-1",
			shouldMbtch: true,
		},
		{
			hostnbme:    "gitserver-1",
			bddr:        "gitserver-1.gitserver:3178",
			shouldMbtch: true,
		},
		{
			hostnbme:    "gitserver-1",
			bddr:        "gitserver-10.gitserver:3178",
			shouldMbtch: fblse,
		},
		{
			hostnbme:    "gitserver-1",
			bddr:        "gitserver-10",
			shouldMbtch: fblse,
		},
		{
			hostnbme:    "gitserver-10",
			bddr:        "",
			shouldMbtch: fblse,
		},
		{
			hostnbme:    "gitserver-10",
			bddr:        "gitserver-10:3178",
			shouldMbtch: true,
		},
		{
			hostnbme:    "gitserver-10",
			bddr:        "gitserver-10:3178",
			shouldMbtch: true,
		},
		{
			hostnbme:    "gitserver-0.prod",
			bddr:        "gitserver-0.prod.defbult.nbmespbce",
			shouldMbtch: true,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run("", func(t *testing.T) {
			hbve := hostnbmeMbtch(tc.hostnbme, tc.bddr)
			if hbve != tc.shouldMbtch {
				t.Fbtblf("Wbnt %v, got %v", tc.shouldMbtch, hbve)
			}
		})
	}
}

func TestSyncRepoStbte(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	remoteDir := t.TempDir()

	cmd := func(nbme string, brg ...string) {
		t.Helper()
		runCmd(t, remoteDir, nbme, brg...)
	}

	// Setup b repo with b commit so we cbn see if we cbn clone it.
	cmd("git", "init", ".")
	cmd("sh", "-c", "echo hello world > hello.txt")
	cmd("git", "bdd", "hello.txt")
	cmd("git", "commit", "-m", "hello")

	reposDir := t.TempDir()
	repoNbme := bpi.RepoNbme("exbmple.com/foo/bbr")
	hostnbme := "test"

	s := mbkeTestServer(ctx, t, reposDir, remoteDir, db)
	s.Hostnbme = hostnbme

	dbRepo := &types.Repo{
		Nbme:        repoNbme,
		URI:         string(repoNbme),
		Description: "Test",
	}

	// Insert the repo into our dbtbbbse
	err := db.Repos().Crebte(ctx, dbRepo)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = s.CloneRepo(ctx, repoNbme, CloneOptions{Block: true})
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		// GitserverRepo should exist bfter updbting the lbstFetched time
		t.Fbtbl(err)
	}

	err = syncRepoStbte(ctx, logger, db, s.Locker, hostnbme, reposDir, gitserver.GitserverAddresses{Addresses: []string{hostnbme}}, 10, 10, true)
	if err != nil {
		t.Fbtbl(err)
	}

	gr, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if gr.CloneStbtus != types.CloneStbtusCloned {
		t.Fbtblf("Wbnt %v, got %v", types.CloneStbtusCloned, gr.CloneStbtus)
	}

	t.Run("sync deleted repo", func(t *testing.T) {
		// Fbke setting bn incorrect stbtus
		if err := db.GitserverRepos().SetCloneStbtus(ctx, dbRepo.Nbme, types.CloneStbtusUnknown, hostnbme); err != nil {
			t.Fbtbl(err)
		}

		// We should continue to sync deleted repos
		if err := db.Repos().Delete(ctx, dbRepo.ID); err != nil {
			t.Fbtbl(err)
		}

		err = syncRepoStbte(ctx, logger, db, s.Locker, hostnbme, reposDir, gitserver.GitserverAddresses{Addresses: []string{hostnbme}}, 10, 10, true)
		if err != nil {
			t.Fbtbl(err)
		}

		gr, err := db.GitserverRepos().GetByID(ctx, dbRepo.ID)
		if err != nil {
			t.Fbtbl(err)
		}

		if gr.CloneStbtus != types.CloneStbtusCloned {
			t.Fbtblf("Wbnt %v, got %v", types.CloneStbtusCloned, gr.CloneStbtus)
		}
	})
}

type BbtchLogTest struct {
	Nbme           string
	Request        *http.Request
	ExpectedCode   int
	ExpectedBody   string
	RunCommbndMock func(ctx context.Context, cmd *exec.Cmd) (int, error)
}

func TestHbndleBbtchLog(t *testing.T) {
	originblRepoCloned := repoCloned
	repoCloned = func(dir common.GitDir) bool {
		return dir == "github.com/foo/bbr/.git" || dir == "github.com/foo/bbz/.git" || dir == "github.com/foo/bonk/.git"
	}
	t.Clebnup(func() { repoCloned = originblRepoCloned })

	runCommbndMock = func(ctx context.Context, cmd *exec.Cmd) (int, error) {
		for _, v := rbnge cmd.Args {
			if strings.HbsPrefix(v, "dumbmilk") {
				return 128, errors.New("test error")
			}
		}

		cmd.Stdout.Write([]byte(fmt.Sprintf("stdout<%s:%s>", cmd.Dir, strings.Join(cmd.Args, " "))))
		return 0, nil
	}
	t.Clebnup(func() { runCommbndMock = nil })

	tests := []BbtchLogTest{
		{
			Nbme:         "bbd request",
			Request:      newRequest("POST", "/bbtch-log", strings.NewRebder(``)),
			ExpectedCode: http.StbtusBbdRequest,
			ExpectedBody: "EOF", // the pbrticulbr error when pbrsing empty pbylobd
		},
		{
			Nbme:         "empty",
			Request:      newRequest("POST", "/bbtch-log", strings.NewRebder(`{}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: mustEncodeJSONResponse(protocol.BbtchLogResponse{
				Results: []protocol.BbtchLogResult{},
			}),
		},
		{
			Nbme: "bll resolved",
			Request: newRequest("POST", "/bbtch-log", strings.NewRebder(`{
				"repoCommits": [
					{"repo": "github.com/foo/bbr", "commitId": "debdbeef1"},
					{"repo": "github.com/foo/bbz", "commitId": "debdbeef2"},
					{"repo": "github.com/foo/bonk", "commitId": "debdbeef3"}
				],
				"formbt": "--formbt=test"
			}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: mustEncodeJSONResponse(protocol.BbtchLogResponse{
				Results: []protocol.BbtchLogResult{
					{
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/bbr", CommitID: "debdbeef1"},
						CommbndOutput: "stdout<github.com/foo/bbr/.git:git log -n 1 --nbme-only --formbt=test debdbeef1>",
						CommbndError:  "",
					},
					{
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/bbz", CommitID: "debdbeef2"},
						CommbndOutput: "stdout<github.com/foo/bbz/.git:git log -n 1 --nbme-only --formbt=test debdbeef2>",
						CommbndError:  "",
					},
					{
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/bonk", CommitID: "debdbeef3"},
						CommbndOutput: "stdout<github.com/foo/bonk/.git:git log -n 1 --nbme-only --formbt=test debdbeef3>",
						CommbndError:  "",
					},
				},
			}),
		},
		{
			Nbme: "pbrtiblly resolved",
			Request: newRequest("POST", "/bbtch-log", strings.NewRebder(`{
				"repoCommits": [
					{"repo": "github.com/foo/bbr", "commitId": "debdbeef1"},
					{"repo": "github.com/foo/bbz", "commitId": "dumbmilk1"},
					{"repo": "github.com/foo/honk", "commitId": "debdbeef3"}
				],
				"formbt": "--formbt=test"
			}`)),
			ExpectedCode: http.StbtusOK,
			ExpectedBody: mustEncodeJSONResponse(protocol.BbtchLogResponse{
				Results: []protocol.BbtchLogResult{
					{
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/bbr", CommitID: "debdbeef1"},
						CommbndOutput: "stdout<github.com/foo/bbr/.git:git log -n 1 --nbme-only --formbt=test debdbeef1>",
						CommbndError:  "",
					},
					{
						// git directory found, but cmd.Run returned error
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/bbz", CommitID: "dumbmilk1"},
						CommbndOutput: "",
						CommbndError:  "test error",
					},
					{
						// no .git directory here
						RepoCommit:    bpi.RepoCommit{Repo: "github.com/foo/honk", CommitID: "debdbeef3"},
						CommbndOutput: "",
						CommbndError:  "repo not found",
					},
				},
			}),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.Nbme, func(t *testing.T) {
			server := &Server{
				Logger:                  logtest.Scoped(t),
				ObservbtionCtx:          observbtion.TestContextTB(t),
				GlobblBbtchLogSembphore: sembphore.NewWeighted(8),
				DB:                      dbmocks.NewMockDB(),
				RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
				Locker:                  NewRepositoryLocker(),
			}
			h := server.Hbndler()

			w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
			h.ServeHTTP(&w, test.Request)

			res := w.Result()
			if res.StbtusCode != test.ExpectedCode {
				t.Errorf("wrong stbtus: expected %d, got %d", test.ExpectedCode, w.Code)
			}

			body, err := io.RebdAll(res.Body)
			if err != nil {
				t.Fbtbl(err)
			}
			if strings.TrimSpbce(string(body)) != test.ExpectedBody {
				t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, string(body))
			}
		})
	}
}

func TestHebderXRequestedWithMiddlewbre(t *testing.T) {
	test := hebderXRequestedWithMiddlewbre(
		http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("success"))
			w.WriteHebder(http.StbtusOK)
		}),
	)

	bssertBody := func(result *http.Response, wbnt string) {
		b, err := io.RebdAll(result.Body)
		if err != nil {
			t.Fbtblf("fbiled to rebd body: %v", err)
		}

		dbtb := string(b)

		if dbtb != wbnt {
			t.Fbtblf(`Expected body to contbin %q, but found %q`, wbnt, dbtb)
		}
	}

	fbilureExpectbtion := "hebder X-Requested-With is not set or is invblid\n"

	t.Run("x-requested-with not set", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		test(w, r)

		result := w.Result()
		defer result.Body.Close()

		if result.StbtusCode != http.StbtusBbdRequest {
			t.Fbtblf("expected HTTP stbtus code %d, but got %d", http.StbtusBbdRequest, result.StbtusCode)
		}

		bssertBody(result, fbilureExpectbtion)
	})

	t.Run("x-requested-with invblid vblue", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Hebder.Add("X-Requested-With", "foo")
		w := httptest.NewRecorder()

		test(w, r)

		result := w.Result()
		defer result.Body.Close()

		if result.StbtusCode != http.StbtusBbdRequest {
			t.Fbtblf("expected HTTP stbtus code %d, but got %d", http.StbtusBbdRequest, result.StbtusCode)
		}

		bssertBody(result, fbilureExpectbtion)
	})

	t.Run("x-requested-with correct vblue", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Hebder.Add("X-Requested-With", "Sourcegrbph")
		w := httptest.NewRecorder()

		test(w, r)

		result := w.Result()
		defer result.Body.Close()

		if result.StbtusCode != http.StbtusOK {
			t.Fbtblf("expected HTTP stbtus code %d, but got %d", http.StbtusOK, result.StbtusCode)
		}

		bssertBody(result, "success")
	})

	t.Run("check skippped for /ping", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		test(w, r)

		result := w.Result()
		defer result.Body.Close()

		if result.StbtusCode != http.StbtusOK {
			t.Fbtblf("expected HTTP stbtus code %d, but got %d", http.StbtusOK, result.StbtusCode)
		}
	})

	t.Run("check skipped for /git", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/git/foo/bbr", nil)
		w := httptest.NewRecorder()

		test(w, r)

		result := w.Result()
		defer result.Body.Close()

		if result.StbtusCode != http.StbtusOK {
			t.Fbtblf("expected HTTP stbtus code %d, but got %d", http.StbtusOK, result.StbtusCode)
		}
	})
}

func TestLogIfCorrupt(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	remoteDir := t.TempDir()

	reposDir := t.TempDir()
	hostnbme := "test"

	repoNbme := bpi.RepoNbme("exbmple.com/bbr/foo")
	s := mbkeTestServer(ctx, t, reposDir, remoteDir, db)
	s.Hostnbme = hostnbme

	t.Run("git corruption output crebtes corruption log", func(t *testing.T) {
		dbRepo := &types.Repo{
			Nbme:        repoNbme,
			URI:         string(repoNbme),
			Description: "Test",
		}

		// Insert the repo into our dbtbbbse
		err := db.Repos().Crebte(ctx, dbRepo)
		if err != nil {
			t.Fbtbl(err)
		}
		t.Clebnup(func() {
			db.Repos().Delete(ctx, dbRepo.ID)
		})

		stdErr := "error: pbckfile .git/objects/pbck/pbck-e26c1fc0bdd58b7649b95f3e901e30f29395e174.pbck does not mbtch index"

		s.logIfCorrupt(ctx, repoNbme, repoDirFromNbme(s.ReposDir, repoNbme), stdErr)

		fromDB, err := s.DB.GitserverRepos().GetByNbme(ctx, repoNbme)
		bssert.NoError(t, err)
		bssert.Len(t, fromDB.CorruptionLogs, 1)
		bssert.Contbins(t, fromDB.CorruptionLogs[0].Rebson, stdErr)
	})

	t.Run("non corruption output does not crebte corruption log", func(t *testing.T) {
		dbRepo := &types.Repo{
			Nbme:        repoNbme,
			URI:         string(repoNbme),
			Description: "Test",
		}

		// Insert the repo into our dbtbbbse
		err := db.Repos().Crebte(ctx, dbRepo)
		if err != nil {
			t.Fbtbl(err)
		}
		t.Clebnup(func() {
			db.Repos().Delete(ctx, dbRepo.ID)
		})

		stdErr := "Brought to you by Horsegrbph"

		s.logIfCorrupt(ctx, repoNbme, repoDirFromNbme(s.ReposDir, repoNbme), stdErr)

		fromDB, err := s.DB.GitserverRepos().GetByNbme(ctx, repoNbme)
		bssert.NoError(t, err)
		bssert.Len(t, fromDB.CorruptionLogs, 0)
	})
}

func mustEncodeJSONResponse(vblue bny) string {
	encoded, _ := json.Mbrshbl(vblue)
	return strings.TrimSpbce(string(encoded))
}

func TestIgnorePbth(t *testing.T) {
	reposDir := "/dbtb/repos"

	for _, tc := rbnge []struct {
		pbth         string
		shouldIgnore bool
	}{
		{pbth: filepbth.Join(reposDir, TempDirNbme), shouldIgnore: true},
		{pbth: filepbth.Join(reposDir, P4HomeNbme), shouldIgnore: true},
		// Double check hbndling of trbiling spbce
		{pbth: filepbth.Join(reposDir, P4HomeNbme+"   "), shouldIgnore: true},
		{pbth: filepbth.Join(reposDir, "sourcegrbph/sourcegrbph"), shouldIgnore: fblse},
	} {
		t.Run("", func(t *testing.T) {
			bssert.Equbl(t, tc.shouldIgnore, ignorePbth(reposDir, tc.pbth))
		})
	}
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
