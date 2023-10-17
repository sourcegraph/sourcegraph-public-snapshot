package outbound

import (
	"context"
	"net"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestEnqueueWebhook(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`"TEST"`)

	t.Run("store error", func(t *testing.T) {
		want := errors.New("mock error")
		store := dbmocks.NewMockOutboundWebhookJobStore()
		store.CreateFunc.SetDefaultReturn(nil, want)
		svc := &outboundWebhookService{store}

		have := svc.Enqueue(ctx, "type", nil, payload)
		assert.ErrorIs(t, have, want)
		mockassert.CalledOnce(t, store.CreateFunc)
	})

	t.Run("success", func(t *testing.T) {
		store := dbmocks.NewMockOutboundWebhookJobStore()
		store.CreateFunc.SetDefaultReturn(&types.OutboundWebhookJob{}, nil)
		svc := &outboundWebhookService{store}

		err := svc.Enqueue(ctx, "type", nil, payload)
		assert.NoError(t, err)
		mockassert.CalledOnce(t, store.CreateFunc)
	})
}

func TestCheckAddress(t *testing.T) {
	t.Run("Invalid Addresses", func(t *testing.T) {
		badURLS := []string{
			// No scheme
			"hi/there?",
			"lol.com",
			"/some/relative/path",
			// Invalid scheme
			"ssh://blah",
			// No host
			"http://",
			// Loopback
			"http://localhost:3000",
			"127.0.0.1",
			"::1",
			// Unspecificed IP
			string(net.IPv4zero),
			string(net.IPv6zero),
			// Private IP
			"10.0.0.0",
			"192.168.255.255",
			"fd00::1",
			// Link-local IP
			"169.254.0.0",
			// Reserved TLD
			"http://somesite.local/some-endpoint",
			"https://somesite.test/some-endpoint",
		}

		for _, badURL := range badURLS {
			err := CheckAddress(badURL)
			if !assert.Error(t, err) {
				t.Fatalf("expected error, got nil for url '%v'", badURL)
			}
		}
	})

	t.Run("Valid Addresses", func(t *testing.T) {
		goodURLS := []string{
			"http://somesite.com/some-endpoint",
			"https://my.webhooks.site/receiver",
			"https://my.webhooks.site:3000/receiver",
			"1.2.3.4",
			"1.2.3.4:2000",
			"2001:0db8:0000:0000:0000:8a2e:0370:7334",
			"2001:db8::8a2e:370:7334",
		}

		for _, goodURL := range goodURLS {
			err := CheckAddress(goodURL)
			if err != nil {
				t.Fatalf("expected nil, got err for url '%v': %v", goodURL, err)
			}
		}
	})
}
