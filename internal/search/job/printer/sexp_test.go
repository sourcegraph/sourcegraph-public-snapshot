package printer

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	. "github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
)

func TestSexp(t *testing.T) {
	autogold.Want("simple sexp", "(TIMEOUT (timeout . 50ms) (AND NoopJob NoopJob))").Equal(t, Sexp(
		NewTimeoutJob(
			50*1_000_000,
			NewAndJob(
				NewNoopJob(),
				NewNoopJob()))))

	autogold.Want("pretty sexp exhaustive cases", `
(SUBREPOPERMSFILTER
  (LIMIT
    (limit . 100)
    (TIMEOUT
      (timeout . 50ms)
      (PARALLEL
        (AND
          NoopJob
          NoopJob)
        (OR
          NoopJob
          NoopJob)
        (AND
          NoopJob
          NoopJob)))))
`).Equal(t, fmt.Sprintf("\n%s\n", PrettySexp(
		NewFilterJob(
			NewLimitJob(
				100,
				NewTimeoutJob(
					50*1_000_000,
					NewParallelJob(
						NewAndJob(
							NewNoopJob(),
							NewNoopJob()),
						NewOrJob(
							NewNoopJob(),
							NewNoopJob()),
						NewAndJob(
							NewNoopJob(),
							NewNoopJob()))))))))
}
