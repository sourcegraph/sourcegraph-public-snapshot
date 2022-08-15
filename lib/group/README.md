# README

## Background

This section offers some explanations below, but it's recommended to skip to the examples table below to get a feel for how this package works.

The main entrypoints to the package are the constructors:

- `group.New()`: A simple Group, equivalent to `go func() + wg.Add()/wg.Done()`
- `group.NewWithResults[T]()`: A group (`ResultGroup`) of tasks that returns results
- `group.NewWithStreaming[T]()`: A group (`StreamGroup`) of tasks that stream their results (in order) with a callback

All of these group types share a set of configuration methods, but some configuration methods are only available for certain types. The following applies to all group types:

- `g.WithMaxConcurrent(n)` returns a group configured to only run up to `n` concurrent goroutines at a time.
- `g.WithConcurrencyLimiter(l)` returns a group configured to use the given limiter. This interface is implemented by the limiter in the `internal/mutablelimiter` package.
- `g.WithErrors()` returns a group (`Error(Result|Stream)?Group`) which runs tasks that return an error.
- `g.WithContext()` returns a group (`Context(Result|Stream)?Group`) that runs tasks that require a context. It will use its context to unblock waiting on the limiter if the context is canceled.

The following is only available after `g.WithErrors()` or `g.WithContext()`:

- `g.WithFirstError()` configures the group to only hold on to the first error returned by a task. By default, it will return a combined error with every error returned by the task. This option is useful in case you don't really need a combined error with a million context errors.

The following is only available after g.WithContext():

- `g.WithCancelOnError()` configures the group to cancel its context if a task returns an error.

This interface makes a few common problems more difficult to hit. Many of these are easy to catch and easy to fix, but if it's even better if they're not possible to hit in the first place. Some common issues this solves are:

- Returning early when a semaphore errors. This means we never clean up the started goroutines, so we return from the function with leaked goroutines.
- Forgetting a `wg.Add()` or `wg.Done()` or `wg.Wait()` (although you still need `g.Wait()`)
- Unintentionally canceling the context on error when using an `errgroup.Group`
- Unintentionally swallowing errors because `errgroup.Group` only returns the first error
- Forgetting mutexes when collecting results or errors
- Not adding a panic handler for goroutines, causing the whole process to crash if they panic (a future version may allow adding a custom panic handler)

Generics are only used where necessary, and the basic `Group/ErrorGroup/ContextGroup` avoids them entirely.

## Usage examples

<table>
<tr>
<th>Without <code>lib/group</code></th>
<th>With <code>lib/group</code></th>
</tr>

<tr>
<td>

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		println("test")
	}()
}
wg.Wait()
```

</td>

<td>

```go
g := group.New()
for i := 0; i < 10; i++ {
	g.Go(func() {
		println("test")
	})
}
g.Wait()
```

</td>
</tr>

<tr>
<td>

```go
var wg sync.WaitGroup
sema := make(chan struct{}, 8)
for i := 0; i < 10; i++ {
	wg.Add(1)
	sema <- struct{}{}
	go func() {
		defer wg.Done()
		defer func() { <-sema }()
		println("test")
	}()
}
wg.Wait()
```

</td>
<td>

```go
g := group.New().WithMaxConcurrency(8)
for i := 0; i < 10; i++ {
	g.Go(func() {
		println("test")
	})
}
g.Wait()
```

</td>

</tr>

<tr>
<td>

```go
g, ctx := errgroup.WithContext(ctx)
sem := semaphore.NewWeighted(int64(32))
for _, item := range chunk {
	item := item
	if err := sem.Acquire(ctx, 1); err != nil {
		return errors.Append(err, g.Wait())
	}

	g.Go(func() error {
		defer sem.Release(1)
		return updateItem(ctx, item)
	})
}
err := g.Wait()
```

</td>
<td>

```go
g := group.New().WithContext(ctx).WithMaxConcurrency(32)
for _, item := range chunk {
	item := item
	g.Go(func(ctx context.Context) error {
		return updateItem(ctx, item)
	})
}
err := g.Wait()
```

</td>
</tr>
<tr>
<td>

```go
var (
	wg sync.WaitGroup
	mu sync.Mutex
	results []int
	errs error
)
for _, name := range names {
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := db.CountNames(ctx, name)
		mu.Lock()
		if err != nil {
			errs = errors.Append(errs, err)
			return
		}
		results = append(results, count)
		mu.Unlock()
	}()
}
wg.Wait()
return results, errs
```

</td>
<td>

```go
g := NewWithResults[int]().WithContext(ctx)
for _, name := range names {
	g.Go(func(ctx context.Context) (int, error) {
		return db.CountNames(ctx)
	})
}
return g.Wait()
```

</tr>
<tr>
<td>

A [surprising amount of code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@ba9bffe9bf5d30a9b6d2e1a764d42162286906d2/-/blob/internal/gitserver/search/search.go?L100-248) that is difficult to understand and difficult to maintain.

</td>
<td>

```go
g := NewWithStreaming[int]().WithContext(ctx).WithMaxConcurrency(8)
callback := func(i int, err error) {
	if err != nil {
		// This will be called in the same order
		// nameStream yields names!
		countStream.Send(i)
	}
}
for nameStream.Next() {
	name := nameStream.Value()
	g.Go(func(ctx context.Context) (int, error) {
		db.CountNames()
	}, cb)
}
g.Wait()
```

</td>
</tr>
</table>
