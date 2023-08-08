package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGuardrails(t *testing.T) {
	// This test is just asserting that our interface is correct. It seems
	// graphql-go only does the schema check if your interface is non-nil.
	_, err := NewSchema(nil, nil, []OptionalResolver{{GuardrailsResolver: guardrailsFake{}}})
	if err != nil {
		t.Fatal(err)
	}
}

type guardrailsFake struct{}

func (guardrailsFake) SnippetAttribution(context.Context, *SnippetAttributionArgs) (SnippetAttributionConnectionResolver, error) {
	return nil, errors.New("fake")
}
