package ipallowlist

import (
	"bytes"
	_ "embed"
	"net/http"
	"text/template"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/ipallowlist/checker"
)

var (
	//go:embed error.html
	errorHTML     string
	errorHTMLTmpl = template.Must(template.New("").Parse(errorHTML))
)

func New(logger log.Logger) *Middleware {
	c := checker.New(logger)

	return &Middleware{
		checker: c,
	}
}

type Middleware struct {
	checker *checker.Checker
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := m.checker.IsAuthorized(r); err != nil {
			w.WriteHeader(http.StatusForbidden)
			data := &pageError{
				StatusCode: http.StatusForbidden,
				StatusText: http.StatusText(http.StatusForbidden),
				Error:      err.Error(),
			}
			var buf bytes.Buffer
			if err := errorHTMLTmpl.Execute(&buf, data); err != nil {
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			_, err = buf.WriteTo(w)
			if err != nil {
				_, _ = w.Write([]byte(err.Error()))
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

type pageError struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
}
