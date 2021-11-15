package shared

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/webhooks"
)

var builtins = map[string]job.Job{
	"webhook-log-janitor": webhooks.NewJanitor(),
}
