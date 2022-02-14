package job

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"
)

func TestSexp(t *testing.T) {
	autogold.Want("simple sexp", "(TIMEOUT 50ms (AND NoopJob NoopJob))").Equal(t, Sexp(
		NewTimeoutJob(
			50*1_000_000,
			NewAndJob(
				NewNoopJob(),
				NewNoopJob()))))

	autogold.Want("pretty sexp exhaustive cases", `
(FILTER
  SubRepoPermissions
  (LIMIT
    100
    (TIMEOUT
      50ms
      (PARALLEL
        (PRIORITY
          (REQUIRED
            (AND
              NoopJob
              NoopJob))
          (OPTIONAL
            (OR
              NoopJob
              NoopJob)))
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
						NewPriorityJob(
							NewAndJob(
								NewNoopJob(),
								NewNoopJob()),
							NewOrJob(
								NewNoopJob(),
								NewNoopJob()),
						),
						NewAndJob(
							NewNoopJob(),
							NewNoopJob()))))))))
}
