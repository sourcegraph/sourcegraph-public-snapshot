package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type observableJob interface {
	Name() string
}

func StartSpan(ctx context.Context, job observableJob) (*trace.Trace, context.Context) {
	tr, ctx := trace.New(ctx, job.Name(), "")
	return tr, ctx
}

func FinishSpan(tr *trace.Trace, alert *search.Alert, err error) {
	tr.SetError(err)
	if alert != nil {
		tr.TagFields(log.String("alert", alert.Title))
	}
	tr.Finish()
}
