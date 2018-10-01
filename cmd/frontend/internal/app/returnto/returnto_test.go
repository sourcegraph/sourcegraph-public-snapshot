package returnto

import (
	"net/http"
	"net/url"
	"testing"
)

func TestURLFromRequest(t *testing.T) {
	tests := []struct {
		url     string
		want    string
		wantErr bool
	}{
		{url: "", want: "/"},
		{url: "?return-to=foo", wantErr: true},
		{url: "?return-to=foo/bar", wantErr: true},
		{url: "?return-to=/foo/bar", want: "/foo/bar"},
		{url: "?return-to=/foo/bar%3Fa=b", want: "/foo/bar?a=b"},
		{url: "?return-to=/foo/bar%3Freturn-to=b", want: "/foo/bar"},
		{url: "?return-to=http://foo", wantErr: true},
		{url: "?return-to=https://foo", wantErr: true},
		{url: "?return-to=//foo", wantErr: true},
	}

	for _, test := range tests {
		u, err := url.Parse(test.url)
		if err != nil {
			t.Error(err)
			continue
		}
		d, err := URLFromRequest(&http.Request{URL: u}, "return-to")
		if (err != nil) != test.wantErr {
			t.Errorf("%s: got err %v, want error? %v", test.url, err, test.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if d.String() != test.want {
			t.Errorf("%s: got %q, want %q", test.url, d, test.want)
		}
	}
}
