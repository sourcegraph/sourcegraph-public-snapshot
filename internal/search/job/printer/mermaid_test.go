package printer

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	. "github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
)

func TestPrettyMermaid(t *testing.T) {
	autogold.Want("simple mermaid", `
flowchart TB
0[AND]
  0---1
  1[NOOP]
    0---2
  2[NOOP]
    `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
		NewAndJob(
			NewNoopJob(),
			NewNoopJob()))))

	autogold.Want("big mermaid", `
flowchart TB
0[SUBREPOPERMSFILTER]
  0---1
  1[LIMIT <br> limit: 100]
    1---2
    2[TIMEOUT <br> timeout: 50ms]
      2---3
      3[PARALLEL]
        3---4
        4[AND]
          4---5
          5[NOOP]
            4---6
          6[NOOP]
            3---7
        7[OR]
          7---8
          8[NOOP]
            7---9
          9[NOOP]
            3---10
        10[AND]
          10---11
          11[NOOP]
            10---12
          12[NOOP]
            `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
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
