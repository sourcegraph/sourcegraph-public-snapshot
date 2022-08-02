package repos

import (
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
		Config: marshalJSON(t, &schema.GitoliteConnection{}),
	}

	_, err := NewGitoliteSource(svc, cf)
	if err != nil {
		t.Fatal(err)
	}
}
