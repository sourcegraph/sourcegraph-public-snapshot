package graphqlbackend

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type externalServiceResolver struct {
	externalService *types.ExternalService
	warning         string

	webhookURLOnce sync.Once
	webhookURL     string
	webhookErr     error
}

const externalServiceIDKind = "ExternalService"

func externalServiceByID(ctx context.Context, id graphql.ID) (*externalServiceResolver, error) {
	// 🚨 SECURITY: Only site admins are allowed to read external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	externalServiceID, err := unmarshalExternalServiceID(id)
	if err != nil {
		return nil, err
	}

	externalService, err := db.ExternalServices.GetByID(ctx, externalServiceID)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{externalService: externalService}, nil
}

func marshalExternalServiceID(id int64) graphql.ID {
	return relay.MarshalID(externalServiceIDKind, id)
}

func unmarshalExternalServiceID(id graphql.ID) (externalServiceID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != externalServiceIDKind {
		err = fmt.Errorf("expected graphql ID to have kind %q; got %q", externalServiceIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &externalServiceID)
	return
}

func (r *externalServiceResolver) ID() graphql.ID {
	return marshalExternalServiceID(r.externalService.ID)
}

func (r *externalServiceResolver) Kind() string {
	return r.externalService.Kind
}

func (r *externalServiceResolver) DisplayName() string {
	return r.externalService.DisplayName
}

func (r *externalServiceResolver) Config() JSONCString {
	return JSONCString(r.externalService.Config)
}

func (r *externalServiceResolver) CreatedAt() DateTime {
	return DateTime{Time: r.externalService.CreatedAt}
}

func (r *externalServiceResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.externalService.UpdatedAt}
}

func (r *externalServiceResolver) WebhookURL() (*string, error) {
	r.webhookURLOnce.Do(func() {
		parsed, err := extsvc.ParseConfig(r.externalService.Kind, r.externalService.Config)
		if err != nil {
			r.webhookErr = errors.Wrap(err, "parsing external service config")
			return
		}
		u := extsvc.WebhookURL(r.externalService.Kind, r.externalService.ID, conf.ExternalURL())
		switch c := parsed.(type) {
		case *schema.BitbucketServerConnection:
			if c.Webhooks != nil {
				r.webhookURL = u
			}
			if c.Plugin != nil && c.Plugin.Webhooks != nil {
				r.webhookURL = u
			}
		case *schema.GitHubConnection:
			if len(c.Webhooks) > 0 {
				r.webhookURL = u
			}
		case *schema.GitLabConnection:
			if len(c.Webhooks) > 0 {
				r.webhookURL = u
			}
		}
	})
	if r.webhookURL == "" {
		return nil, r.webhookErr
	}
	return &r.webhookURL, r.webhookErr
}

func (r *externalServiceResolver) Warning() *string {
	if r.warning == "" {
		return nil
	}
	return &r.warning
}
