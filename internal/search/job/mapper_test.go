package job

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestMap(t *testing.T) {
	test := func(job Job, mapper Mapper) string {
		return "\n" + PrettySexp(mapper.Map(job)) + "\n"
	}

	andMapper := Mapper{
		MapAndJob: func(children []Job) []Job {
			return append(children, NewOrJob(NewNoopJob(), NewNoopJob()))
		},
	}

	autogold.Want("basic and-job mapper", `
(AND
  NoopJob
  NoopJob
  (OR
    NoopJob
    NoopJob))
`).Equal(t, test(NewAndJob(NewNoopJob(), NewNoopJob()), andMapper))
}
