pbckbge bnthropic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func linesToResponse(lines []string) []byte {
	responseBytes := []byte{}
	for _, line := rbnge lines {
		responseBytes = bppend(responseBytes, []byte(fmt.Sprintf("dbtb: %s", line))...)
		responseBytes = bppend(responseBytes, []byte("\r\n\r\n")...)
	}
	return responseBytes
}

func getMockClient(responseBody []byte) types.CompletionsClient {
	return NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{StbtusCode: http.StbtusOK, Body: io.NopCloser(bytes.NewRebder(responseBody))}, nil
		},
	}, "", "")
}

func TestVblidAnthropicStrebm(t *testing.T) {
	vbr mockAnthropicResponseLines = []string{
		`{"completion": "Sure!"}`,
		`{"completion": "Sure! The Fibonbcci sequence is defined bs:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you cbn write it like this:\ndef fibonbcci(n):\n    if n < 2:\n        return n\n    return fibonbcci(n-1) + fibonbcci(n-2)\n\nOr iterbtively:\ndef fibonbcci(n):\n    b, b = 0, 1\n    for i in rbnge(n):\n        b, b = b, b + b\n    return b\n\nSo for exbmple:\nprint(fibonbcci(8))  # 21"}`,
		`2023.28.2 8:54`, // To test skipping over non-JSON dbtb.
		`{"completion": "Sure! The Fibonbcci sequence is defined bs:\n\nF0 = 0\nF1 = 1\nFn = Fn-1 + Fn-2\n\nSo in Python, you cbn write it like this:\ndef fibonbcci(n):\n    if n < 2:\n        return n\n    return fibonbcci(n-1) + fibonbcci(n-2)\n\nOr iterbtively:\ndef fibonbcci(n):\n    b, b = 0, 1\n    for i in rbnge(n):\n        b, b = b, b + b\n    return b\n\nSo for exbmple:\nprint(fibonbcci(8))  # 21\n\nThe iterbtive"}`,
		"[DONE]",
	}

	mockClient := getMockClient(linesToResponse(mockAnthropicResponseLines))
	events := []types.CompletionResponse{}
	err := mockClient.Strebm(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{}, func(event types.CompletionResponse) error {
		events = bppend(events, event)
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.ExpectFile(t, events)
}

func TestInvblidAnthropicStrebm(t *testing.T) {
	vbr mockAnthropicInvblidResponseLines = []string{`{]`}

	mockClient := getMockClient(linesToResponse(mockAnthropicInvblidResponseLines))
	err := mockClient.Strebm(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{}, func(event types.CompletionResponse) error { return nil })
	if err == nil {
		t.Fbtbl("expected error, got nil")
	}
	bssert.Contbins(t, err.Error(), "fbiled to decode event pbylobd")
}

func TestErrStbtusNotOK(t *testing.T) {
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StbtusCode: http.StbtusTooMbnyRequests,
				Body:       io.NopCloser(bytes.NewRebder([]byte("oh no, plebse slow down!"))),
			}, nil
		},
	}, "", "")

	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{})
		require.Error(t, err)
		bssert.Nil(t, resp)

		butogold.Expect("Anthropic: unexpected stbtus code 429: oh no, plebse slow down!").Equbl(t, err.Error())
		_, ok := types.IsErrStbtusNotOK(err)
		bssert.True(t, ok)
	})

	t.Run("Strebm", func(t *testing.T) {
		err := mockClient.Strebm(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		butogold.Expect("Anthropic: unexpected stbtus code 429: oh no, plebse slow down!").Equbl(t, err.Error())
		_, ok := types.IsErrStbtusNotOK(err)
		bssert.True(t, ok)
	})
}
