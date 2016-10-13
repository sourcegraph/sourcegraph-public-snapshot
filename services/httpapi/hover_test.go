package httpapi

import "testing"

func TestHackParseDefKeyPath(t *testing.T) {
	tests := []struct {
		title, want string
	}{
		{
			title: "func NewRouter() *Router",
			want:  "NewRouter",
		},
		{
			title: "type Router struct{NotFoundHandler Handler; parent parentRoute; routes []*Route; namedRoutes map[string]*Route; strictSlash bool; skipClean bool; KeepContext bool; useEncodedPath bool}",
			want:  "Router",
		},
		{
			title: "type Request struct{Method string; URL *URL; Proto string; ProtoMajor int; ProtoMinor int; Header Header; Body ReadCloser; ContentLength int64; TransferEncoding []string; Close bool; Host string; Form Values; PostForm Values; MultipartForm *Form; Trailer Header; RemoteAddr string; RequestURI string; TLS *ConnectionState; Cancel <-chan struct{}; Response *Response; ctx Context}",
			want:  "Request",
		},
		{
			title: "struct field Method string",
			want:  "", // Impossible to derive information for struct fields.
		},
		{
			title: "func (*Router).Match(req *Request, match *RouteMatch) bool",
			want:  "Router/Match",
		},
	}
	for _, tst := range tests {
		t.Run(tst.title, func(t *testing.T) {
			got := hackParseDefKeyPath(tst.title)
			if got != tst.want {
				t.Logf("got %q want %q\n", got, tst.want)
				t.Fail()
			}
		})
	}
}
