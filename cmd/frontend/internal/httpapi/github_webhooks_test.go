package httpapi

import (
	"context"
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"testing"
)

func TestDispatchSuccess(t *testing.T) {
	h := GithubWebhook{}
	var called bool
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called = true
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !called {
		t.Errorf("Expected called to be true, was false")
	}
}

func TestDispatchNoHandler(t *testing.T) {
	h := GithubWebhook{}
	ctx := context.Background()
	// no op
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestDispatchSuccessMultiple(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if called != 2 {
		t.Errorf("Expected called to be 2, got %v", called)
	}
}

func TestDispatchError(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return fmt.Errorf("oh dear")
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); errString(err) != "oh dear" {
		t.Errorf("Expected 'oh no', got %s", err)
	}
	if called != 1 {
		t.Errorf("Expected called to be 1, got %v", called)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
