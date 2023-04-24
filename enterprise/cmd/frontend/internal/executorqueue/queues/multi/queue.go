package multi

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
)

func QueueHandler(handlers map[string]handler.ExecutorHandler) MultiQueueHandler {
	return MultiQueueHandler{Handlers: handlers}
}
