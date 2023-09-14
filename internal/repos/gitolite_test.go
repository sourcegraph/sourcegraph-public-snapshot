package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitoliteSource(t *testing.T) {
	cf, save := newClientFactoryWithOpt(t, "basic", httpcli.ExternalTransportOpt)
	defer save(t)

	svc := &types.ExternalService{
		Kind:   extsvc.KindGitolite,
		Config: extsvc.NewUnencryptedConfig(MarshalJSON(t, &schema.GitoliteConnection{})),
	}

	ctx := context.Background()
	_, err := NewGitoliteSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}
}
