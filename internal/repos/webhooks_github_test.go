package repos

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/joho/godotenv"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var repoName = "ghe.sgdev.org/milton/test"
var updateWebhook = flag.Bool("updateWebhook", false, "updateWebhook github webhooks API testdata")

func TestGitHubWebhooks_List(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("./.env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.Client = NewTestClient(t, "List", updateWebhook)

	_, err = gh.ListSyncWebhooks(ctx, repoName, token)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGitHubWebhooks_CreateListDelete(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("./.env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.Client = NewTestClient(t, "CreateListDelete", updateWebhook)

	id, err := gh.CreateSyncWebhook(ctx, repoName, "https://repoupdater.com", "secret", token)
	if err != nil {
		t.Fatal(err)
	}

	payloads, err := gh.ListSyncWebhooks(ctx, repoName, token)
	if err != nil {
		t.Fatal(err)
	}

	idCount := 0
	for _, payload := range payloads {
		if payload.ID == id {
			idCount++
		}
	}

	// even if there is an error,
	// we want to delete the webhook we just created
	countErr := ""
	if idCount != 1 {
		countErr = "Created webhook more not equal to 1"
	}

	deleted, err := gh.DeleteSyncWebhook(ctx, repoName, id, token)
	if err != nil {
		t.Fatal(err)
	}

	if countErr != "" {
		t.Fatal(countErr)
	}

	if !deleted {
		t.Fatalf("Failed to delete recently created webhook")
	}
}

func TestGitHubWebhooks_Find(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("./.env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.Client = NewTestClient(t, "Find", updateWebhook)

	_, found := gh.FindSyncWebhook(ctx, repoName, token)
	if !found {
		t.Fatalf("Could not find webhook")
	}
}

func TestCreateFile(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("./.env")
	if err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("ACCESS_TOKEN")

	gh, err := NewGithubWebhookAPI()
	if err != nil {
		t.Fatal(err)
	}
	gh.Client = NewTestClient(t, "CreateFile", updateWebhook)

	// url := fmt.Sprintf("%s://%s", globals.	ExternalURL().Scheme, globals.ExternalURL().Host)

	id, err := gh.CreateSyncWebhook(ctx, repoName, "https://a125-116-15-22-254.ap.ngrok.io", "secret", token)
	if err != nil {
		t.Fatal(err)
	}

	success, err := gh.createFile(ctx, repoName, id, token)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("success:", success)
}

func (g GithubWebhookAPI) createFile(ctx context.Context, repoName string, hookID int, token string) (bool, error) {
	u, err := urlBuilderWithID(repoName, hookID)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/tests", u)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := g.Client.do(ctx, req)
	if err != nil {
		return false, err
	}
	fmt.Printf("resp:%+v\n", resp)

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	fmt.Println("Body:", string(bs))

	if resp.StatusCode != 204 {
		return false, errors.Newf("non-204 status code: %d", resp.StatusCode)
	}

	return true, nil
}

func NewTestClient(t testing.TB, name string, updateWebhook *bool) *Client {
	t.Helper()

	casette := filepath.Join("testdata/vcr/", name)
	rec, err := httptestutil.NewRecorder(casette, *updateWebhook)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to updateWebhook test data: %s", err)
		}
	})
	rec.SetMatcher(ignoreHostMatcher)

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	cli, err := NewClient(hc)
	if err != nil {
		t.Fatal(err)
	}

	return cli
}

func ignoreHostMatcher(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}
	u, err := url.Parse(i.URL)
	if err != nil {
		return false
	}
	u.Host = r.URL.Host
	u.Scheme = r.URL.Scheme
	return r.URL.String() == u.String()
}
