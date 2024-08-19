# Glock

[![GoDoc](https://godoc.org/github.com/efritz/glock?status.svg)](https://godoc.org/github.com/efritz/glock)
[![Build Status](https://secure.travis-ci.org/efritz/glock.png)](http://travis-ci.org/efritz/glock)
[![Maintainability](https://api.codeclimate.com/v1/badges/45c92a2ed058b29a2afc/maintainability)](https://codeclimate.com/github/efritz/glock/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/45c92a2ed058b29a2afc/test_coverage)](https://codeclimate.com/github/efritz/glock/test_coverage)

Small go library for mocking parts of the [time](https://golang.org/pkg/time) and [context](https://golang.org/pkg/context) packages.

## Time Utilities

The package contains a `Clock` and `Ticker` interface that wrap the `time.Now`, `time.After`, and `time.Sleep` functions and the `Ticker` struct, respectively.

A *real* clock can be created for general (non-test) use. This implementation simply falls back to the functions provided in the time package.

```go
clock := glock.NewRealClock()
clock.Now()                       // calls time.Now
clock.After(time.Second)          // calls time.After(time.Second)

t := clock.NewTicker(time.Second) // wraps time.NewTicker(time.Second)
t.Chan()                          // returns ticker's C field
t.Stop()                          // stops the ticker
```

In order to make unit tests that depend on time deterministic (and free of sleep calls), a *mock* clock can be used in place of the real clock. The mock clock allows you to control the current time with `SetCurrent` and `Advance` methods.

```go
clock := glock.NewMockClock()

clock.Now() // returns time of creation
clock.Now() // returns time of creation
clock.SetCurrent(time.Unix(603288000, 0))
clock.Now() // returns Feb 12, 1989
clock.Advance(time.Day)
clock.Now() // returns Feb 13, 1989
```

The `Advance` method will also trigger a value on the channels created by the `After` and `Ticker` functions, if enough virtual time has elapsed for the events to fire.

```go
clock := glock.NewMockClockAt(time.Unix(603288000, 0))

c1 := clock.After(time.Second)
c2 := clock.After(time.Minute)
clock.GetAfterArgs()            // returns {time.Second, time.Minute}
clock.GetAfterArgs()            // returns {}
clock.Advance(time.Second * 30) // Fires c2
clock.Advance(time.Second * 30) // Fires c1
```

```go
clock := glock.NewMockClock()

ticker := clock.NewTicker(time.Minute)
defer ticker.Stop()

go func() {
    for range ticker.Chan() {
        // ...
    }
}()

clock.Advance(time.Second * 30)
clock.Advance(time.Second * 30) // Fires ch
clock.Advance(time.Second * 30)
clock.Advance(time.Second * 30) // Fires ch
```

The `Advance` method will send a value to any current listener registered to a channel on the clock. Timing these calls in relation with the clock consumer is not always an easy task. A variation of the advance method, `BlockingAdvance` can be used in its place when you want to first ensure that there is a listener on a channel returned by `After`.


```go
clock := glock.NewMockClock()

go func() {
    <-clock.After(time.Second * 30)
}()

clock.BlockingAdvance(time.Second * 30) // blocks until the concurrent call to After
clock.BlockingAdvance(time.Second * 30) // blocks indefinitely as there are no listeners
```

Ticker instances themselves have the same time advancing mechanisms. Using `Advance` on a ticker (or using `Advance` on the clock from which a ticker was created) will cause the ticker to fire _once_ and then forward itself to the current time. This mimics the behavior of the Go runtime clock (see the test functions `^TestTickerOffset`).

Where the `Advance` method sends the ticker's time to the consumer in a background goroutine, the `BlockingAdvance` variant will send the value in the caller's goroutine.

```go
ticker := clock.NewMockTicker(time.Second * 30)
defer ticker.Stop()

go func() {
    <-ticker.Chan()
    <-ticker.Chan()
    <-ticker.Chan()
}()

ticker.BlockingAdvance(time.Second * 15)
ticker.BlockingAdvance(time.Second * 15) // Fires ch
ticker.BlockingAdvance(time.Second * 15)
ticker.BlockingAdvance(time.Second * 15) // Fires ch
ticker.BlockingAdvance(time.Second * 60) // Fires ch _once_

ticker.Advance(time.Second * 30)         // does not block; sent asynchronously
ticker.BlockingAdvance(time.Second * 30) // blocks indefinitely as there are no listeners
```

## Context Utilities

If you'd like to use a `context.Context` as a way to make a glock `Clock` available, this
package provides `WithContext` and `FromContext` utility methods.

To add a `Clock` to a context you would call `WithContext` and provide a parent context as well
as the `Clock` you'd like to add.

```go
clock := glock.NewMockClock()

ctx := context.Background()
ctx = glock.WithContext(ctx, clock)
```

To retrieve the `Clock` from a context, the `FromContext` method is available. If a `Clock`
does not already exist within the context `FromContext` will return a new *real* clock instance.

```go
// Retrieve a mock clock from the context
clock := glock.NewMockClock()

ctx := context.Background()
ctx = glock.WithContext(ctx, clock)

ctxClock := glock.FromContext(ctx)
```

```go
// Retrieve a default real clock from the context
ctx := context.Background()
ctx = glock.WithContext(ctx, clock)

ctxClock := glock.FromContext(ctx)
```

## Context Testing Utilities

The package also contains the functions `ContextWithDeadline` and `ContextWithTimeout` that
mimic the `context.WithDeadline` and `context.WithTimeout` functions, but will use a
user-provided `Clock` instance rather than the standard `time.After` function.

A *real* clock can be used for non-test scenarios without much additional overhead.

```go
clock := glock.NewRealClock()
ctx, cancel := glock.ContextWithTimeout(context.Background(), clock, time.Second)
defer cancel()

<-ctx.Done() // Waits 1s
```

In order to make unit tests that depend on context timeouts deterministic, a *mock* clock can
be used in place of the real clock. The mock clock can be advanced in the same was a described
in the previous section.

```go
clock := glock.NewMockClock()
ctx, cancel := glock.ContextWithTimeout(context.Background(), clock, time.Second)
defer cancel()

go func() {
    <-time.After(time.Millisecond * 250)
    clock.BlockingAdvance(time.Second)
}()

<-ctx.Done() // Waits around 250ms
```

## License

Copyright (c) 2021 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
