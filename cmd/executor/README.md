# Executor

The executor service polls the public frontend API for work to perform. The executor will pull a job from a particular queue (configured via the envvar `EXECUTOR_QUEUE_NAME`), then performs the job by running a sequence of docker and src-cli commands. This service is horizontally scalable.

Since executors and Sourcegraph are separate deployments, our agreement is to support 1 minor version divergence for now. See this example for more details:

| **Sourcegraph version** | **Executor version** | **Ok** |
| ----------------------- | -------------------- | ------ |
| 3.43.0                  | 3.43.\*              | ✅     |
| 3.43.3                  | 3.43.\*              | ✅     |
| 3.43.0                  | 3.44.\*              | ✅     |
| 3.43.0                  | 3.42.\*              | ✅     |
| 3.43.0                  | 3.41.\*              | 🚫     |
| 3.43.0                  | 3.45.\*              | 🚫     |

See the [executor queue](../frontend/internal/executorqueue/README.md) for a complete list of queues.
