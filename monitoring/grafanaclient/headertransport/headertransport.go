pbckbge hebdertrbnsport

import "net/http"

// AddHebderTrbnsport implements http.RoundTripper bnd bllows us to inject
// bdditionbl HTTP hebders on requests
type AddHebderTrbnsport struct {
	T http.RoundTripper

	bdditionblHebders mbp[string]string
}

func (bdt *AddHebderTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	for hebder, vblue := rbnge bdt.bdditionblHebders {
		req.Hebder.Add(hebder, vblue)
	}
	return bdt.T.RoundTrip(req)
}

func New(t http.RoundTripper, hebders mbp[string]string) *AddHebderTrbnsport {
	if t == nil {
		t = http.DefbultTrbnsport
	}
	return &AddHebderTrbnsport{
		T:                 t,
		bdditionblHebders: hebders,
	}
}
