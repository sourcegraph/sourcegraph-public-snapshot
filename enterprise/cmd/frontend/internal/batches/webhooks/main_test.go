package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"os"
	"testing"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var update = flag.Bool("update", false, "update testdata")

func getSingleRepo(ctx context.Context, bitbucketSource *repos.BitbucketServerSource, name string) (*types.Repo, error) {
	repoChan := make(chan repos.SourceResult)
	go func() {
		bitbucketSource.ListRepos(ctx, repoChan)
		close(repoChan)
	}()

	var bitbucketRepo *types.Repo
	for result := range repoChan {
		if result.Err != nil {
			return nil, result.Err
		}
		if result.Repo == nil {
			continue
		}
		if string(result.Repo.Name) == name {
			bitbucketRepo = result.Repo
		}
	}

	return bitbucketRepo, nil
}

type webhookTestCase struct {
	Payloads []struct {
		PayloadType string          `json:"payload_type"`
		Data        json.RawMessage `json:"data"`
	} `json:"payloads"`
	ChangesetEvents []*btypes.ChangesetEvent `json:"changeset_events"`
}

func loadWebhookTestCase(t testing.TB, path string) webhookTestCase {
	t.Helper()

	bs, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var tc webhookTestCase
	if err := json.Unmarshal(bs, &tc); err != nil {
		t.Fatal(err)
	}
	for i, ev := range tc.ChangesetEvents {
		meta, err := btypes.NewChangesetEventMetadata(ev.Kind)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(ev.Metadata)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(raw, &meta)
		if err != nil {
			t.Fatal(err)
		}
		tc.ChangesetEvents[i].Metadata = meta
	}

	return tc
}

func sign(t *testing.T, message, secret []byte) string {
	t.Helper()

	mac := hmac.New(sha256.New, secret)

	_, err := mac.Write(message)
	if err != nil {
		t.Fatalf("writing hmac message failed: %s", err)
	}

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
