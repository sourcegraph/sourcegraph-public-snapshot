package vcsclient

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"strings"

	"sourcegraph.com/sourcegraph/vcsstore/git"
)

func Test_gitTransport_InfoRefs(t *testing.T) {
	tests := []struct {
		repoPath string
		service  string
		expURL   string
		expOut   string
	}{{
		repoPath: "a.b/c",
		service:  "receive-pack",
		expURL:   "/a.b/c/.git/info/refs?service=git-receive-pack",
		expOut: `0090542272db9b9b8f3dfd57ab143176c9ecaf7f6abb refs/heads/custom-context report-status delete-refs side-band-64k quiet ofs-delta agent=git/1.9.1
003f8096f47503459bcc74d1f4c487b7e6e42e5746b5 refs/heads/master
0000`,
	}, {
		repoPath: "a.b/c",
		service:  "receive-pack",
		expURL:   "/a.b/c/.git/info/refs?service=git-receive-pack",
		expOut: `0090542272db9b9b8f3dfd57ab143176c9ecaf7f6abb refs/heads/custom-context report-status delete-refs side-band-64k quiet ofs-delta agent=git/1.9.1
003f8096f47503459bcc74d1f4c487b7e6e42e5746b5 refs/heads/master
0000`,
	}, {
		repoPath: "a.b/c",
		service:  "upload-pack",
		expURL:   "/a.b/c/.git/info/refs?service=git-upload-pack",
		expOut: `00d18096f47503459bcc74d1f4c487b7e6e42e5746b5 HEADmulti_ack thin-pack side-band side-band-64k ofs-delta shallow no-progress include-tag multi_ack_detailed no-done symref=HEAD:refs/heads/master agent=git/1.9.1
		0047542272db9b9b8f3dfd57ab143176c9ecaf7f6abb refs/heads/custom-context
		003f8096f47503459bcc74d1f4c487b7e6e42e5746b5 refs/heads/master
		0000`,
	}}

	for _, test := range tests {
		func() {
			setup()
			defer teardown()

			expURL, _ := url.Parse(test.expURL)

			gitTransport, err := vcsclient.GitTransport(test.repoPath)
			if err != nil {
				t.Fatal(err)
			}

			var called bool
			mux.HandleFunc(expURL.Path, func(w http.ResponseWriter, r *http.Request) {
				called = true

				if userAgent := r.Header.Get("User-Agent"); !strings.HasPrefix(userAgent, "git/") {
					t.Errorf("expected User-Agent git/*, but got %s", userAgent)
				}

				if r.Method != "GET" {
					t.Errorf("expected GET, got %s", r.Method)
				}

				expQuery, actQuery := r.URL.Query().Encode(), expURL.Query().Encode()
				if expQuery != actQuery {
					t.Errorf("expected URL query %s but got %s", expQuery, actQuery)
				}
				w.Write([]byte(test.expOut))
			})

			var buf bytes.Buffer
			err = gitTransport.InfoRefs(&buf, test.service)
			if err != nil {
				t.Errorf("unexpected error calling gitTransport.InfoRefs: %s", err)
			}

			if !called {
				t.Errorf("endpoint never called")
			}
		}()
	}
}

func Test_gitTransport_ReceivePack(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	opt := git.GitTransportOpt{}
	expURL := "/a.b/c/.git/git-receive-pack"
	expIn := "this is the expected input"
	expOut := "this is the expected output"

	gitTransport, err := vcsclient.GitTransport(repoPath)
	if err != nil {
		t.Fatal(err)
	}

	var called bool

	mux.HandleFunc(expURL, func(w http.ResponseWriter, r *http.Request) {
		called = true

		if userAgent := r.Header.Get("User-Agent"); !strings.HasPrefix(userAgent, "git/") {
			t.Errorf("expected User-Agent git/*, but got %s", userAgent)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		in, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if expIn != string(in) {
			t.Errorf("expected input \"%s\" but got \"%s\"", expIn, string(in))
		}

		w.Write([]byte(expOut))
	})

	var out bytes.Buffer
	in := bytes.NewReader([]byte(expIn))
	err = gitTransport.ReceivePack(&out, in, opt)
	if err != nil {
		t.Fatalf("unexpected error calling gitTransport.ReceivePack: %s", err)
	}

	if !called {
		t.Errorf("endpoint never called")
	}

	if expOut != string(out.Bytes()) {
		t.Errorf("expected output \"%s\" but got \"%s\"", expOut, string(out.Bytes()))
	}
}

func Test_gitTransport_UploadPack(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	opt := git.GitTransportOpt{}
	expURL := "/a.b/c/.git/git-upload-pack"
	expIn := "this is the expected input"
	expOut := "this is the expected output"

	gitTransport, err := vcsclient.GitTransport(repoPath)
	if err != nil {
		t.Fatal(err)
	}

	var called bool

	mux.HandleFunc(expURL, func(w http.ResponseWriter, r *http.Request) {
		called = true

		if userAgent := r.Header.Get("User-Agent"); !strings.HasPrefix(userAgent, "git/") {
			t.Errorf("expected User-Agent git/*, but got %s", userAgent)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		in, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if expIn != string(in) {
			t.Errorf("expected input \"%s\" but got \"%s\"", expIn, string(in))
		}

		w.Write([]byte(expOut))
	})

	var out bytes.Buffer
	in := bytes.NewReader([]byte(expIn))
	err = gitTransport.UploadPack(&out, in, opt)
	if err != nil {
		t.Fatalf("unexpected error calling gitTransport.UploadPack: %s", err)
	}

	if !called {
		t.Errorf("endpoint never called")
	}

	if expOut != string(out.Bytes()) {
		t.Errorf("expected output \"%s\" but got \"%s\"", expOut, string(out.Bytes()))
	}
}
