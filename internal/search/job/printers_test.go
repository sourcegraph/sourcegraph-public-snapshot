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

func TestPrettyMermaid(t *testing.T) {
	autogold.Want("simple mermaid", `
flowchart TB
0([AND])
  0---1
  1([NoopJob])
  0---2
  2([NoopJob])
  `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
		NewAndJob(
			NewNoopJob(),
			NewNoopJob()))))

	autogold.Want("big mermaid", `
flowchart TB
0([FILTER])
  0---1
  1[SubRepoPermissions]
  0---2
  2([LIMIT])
    2---3
    3[100]
    2---4
    4([TIMEOUT])
      4---5
      5[50ms]
      4---6
      6([PARALLEL])
        6---7
        7([PRIORITY])
          7---8
          8([REQUIRED])
          8---9
          9([AND])
            9---10
            10([NoopJob])
            9---11
            11([NoopJob])
            7---12
          12([OPTIONAL])
          12---13
          13([OR])
            13---14
            14([NoopJob])
            13---15
            15([NoopJob])
            6---16
        16([AND])
          16---17
          17([NoopJob])
          16---18
          18([NoopJob])
          `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
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
