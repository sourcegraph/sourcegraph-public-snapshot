package gitlab

import (
	"context"
	"testing"
)

func TestGetVersion(t *testing.T) {
	ctx := context.Background()

	client := newTestClient(t)
	client.httpClient = &mockHTTPResponseBody{
		responseBody: `{"version":"12.7.2-ee","revision":"be1bc017799"}`,
	}

	have, err := client.GetVersion(ctx)
	if err != nil {
		t.Errorf("unexpected non-nil error: %+v", err)
	}

	if want := "12.7.2-ee"; have != want {
		t.Errorf("wrong version. want=%s, have=%s", want, have)
	}
}
