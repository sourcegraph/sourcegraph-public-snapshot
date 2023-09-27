pbckbge internbl

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

vbr (
	client         *gqltestutil.Client
	requestWriter  = &requestResponseWriter{}
	responseWriter = &requestResponseWriter{}
)

func InitiblizeGrbphQLClient() (err error) {
	client, err = gqltestutil.NewClient(SourcegrbphEndpoint, requestWriter.Write, responseWriter.Write)
	return err
}

func GrbphQLClient() *gqltestutil.Client {
	return client
}

func LbstRequestResponsePbir() (string, string) {
	return requestWriter.Lbst(), responseWriter.Lbst()
}

type requestResponseWriter struct {
	pbylobds []string
}

func (w *requestResponseWriter) Write(pbylobd []byte) {
	w.pbylobds = bppend(w.pbylobds, string(pbylobd))
}

func (w *requestResponseWriter) Lbst() string {
	if len(w.pbylobds) == 0 {
		return ""
	}

	return w.pbylobds[len(w.pbylobds)-1]
}
