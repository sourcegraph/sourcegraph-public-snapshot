# Developing a worker

Workers are the consumer of a producer/consumer relationship.

TODO

## Overview

A **worker** is an generic process configured with a _store_ and a _handler_. In short, the store describes how to interact with where jobs are persisted; the handler (supplied by the user describes how to process each job. Both of these components will be discussed in more detail below.

The **store** is responsible for selecting the next available job from the backing persistence layer and suitably _locking_ it from other consumers as well as updating the job records as they make progress in the handler. Generally, this will be an instance of [dbworker/store.Store](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/blob/internal/workerutil/dbworker/store/store.go#L204:6), although there are [other implementations](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/blob/enterprise/cmd/executor/internal/apiclient/client.go#L51:6).

The **handler** is responsible for handling a single job once dequeued from the associated store. Handlers can be fairly minimal, but there are a number of hooks which can be overridden to customize the behavior of the worker.

#### Hook 1: Pre-dequeue (optional)

Before the worker dequeues, the next job, the _pre dequeue_ hook (if defined) is invoked. The hook has the following signature:

```
func (h *myHandler) PreDequeue(context.Context) (dequeueable bool, extraDequeueArguments interface{}, err error) {
  // configure conditional job selection
  return true, nil, nil
}
```

If the hook returns with `dequeueable = false`, the worker will continue to wait before the next attempt to dequeue an available job. If the hook returns with `extraDequeueArguments`, then it will be passed (in an implementation-specific manner) to the store while dequeueing a job. For the database-backed store, the `extraDequeueArguments` take the form of `*sqlf.Query` expressions, which are added to the conditional clause when selecting a candidate job record.

The main use of these return values are to aide in implementation of a worker _budget_. If the worker is processing multiple jobs, it must be careful that the maximum number of concurrent jobs do not exceed the resource capacity of the worker process. Adding additional conditions to the dequeue method in this manner allows us to skip over jobs that would require more resources than our current capacity. (This applies to jobs for which the resource usage can be fairly accurately estimated.)

#### Hook 2: PreHandle (optional)

After the worker dequeues a record to process, but before it's processed, the _pre handle_ hook (if defined) is invoked. The hook has the following signature:

```
func (h *myHandler) PreHandle(ctx context.Context, record Record) {
  // do something before
}
```

The record value is what was dequeued from the backing store. It's type is a nearly useless interface, thus the value must be cast to the expected type of job of concern to this handler before doing anything useful with it.

Along with the `PostHandle` hook described below, these hooks can effectively maintain the worker budget discussed above: before processing each job we atomically decrease our worker's current _headroom_, and restore the headroom once the job has completed.

#### Hook 3: Handle (required)

To process a record, the worker invokes the _handle_ hook, which is the only required hook. The hook has the following signature:

```
func (h *myHandler) Handle(ctx context.Context, store Store, record Record) error {
  // process record
  return nil
}
```

The record value is what was dequeued from the backing store. It's type is a nearly useless interface, thus the value must be cast to the expected type of job of concern to this handler before doing anything useful with it.

The store passed along with the record may be refined version of the store configured with the worker. For the database-backed store, it is a version of the configured store, but has been modified to execute all statements within the transaction that locked the record.

After processing a job, the worker will update a job's state (via the store) according to the handle hook's return value. A nil error will result in a _complete_ job; a retryable error (according to [this function](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/blob/internal/errcode/code.go#L174:6)) will result in an _errored_ job (which may be retried); any other error will result in a _failed_ job (which are not retried).

#### Hook 4: PostHandle (optional)

After the worker processes a record (successfully _or_ unsuccessfully), the _post handle_ hook (if defined) is invoked. The hook has the following signature:

```
func (h *myHandler) PostHandle(ctx context.Context, record Record) {
  // do something after
}
```

The record value is what was just processed. It's type is a nearly useless interface, thus the value must be cast to the expected type of job of concern to this handler before doing anything useful with it.

## Database-backed store

TODO

# ~~~ Notes ~~~

The worker's throughput behavior can be modified by adjusting additional parameters:

- Supply `Interval` to modify the delay between jobs per handler.
- Supply `NumHandlers` to adjust the number of concurrent jobs active per process.
