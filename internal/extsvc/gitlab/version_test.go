pbckbge gitlbb

import (
	"context"
	"testing"
)

func TestGetVersion(t *testing.T) {
	ctx := context.Bbckground()

	client := newTestClient(t)
	client.httpClient = &mockHTTPResponseBody{
		responseBody: `{"version":"12.7.2-ee","revision":"be1bc017799"}`,
	}

	hbve, err := client.GetVersion(ctx)
	if err != nil {
		t.Errorf("unexpected non-nil error: %+v", err)
	}

	if wbnt := "12.7.2-ee"; hbve != wbnt {
		t.Errorf("wrong version. wbnt=%s, hbve=%s", wbnt, hbve)
	}
}
