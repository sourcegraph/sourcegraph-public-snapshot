package slices

import (
	"context"
	"sync"
)

// AllAsync returns true if f returns true for all elements in items.
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
func AllAsync[S ~[]T, T any](items S, workers int, f func(el T) bool) bool {
	if len(items) == 0 {
		return true
	}

	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int, result chan<- bool, ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case index, ok := <-jobs:
				if !ok {
					return
				}
				if !f(items[index]) {
					result <- false
					return
				}
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// when we're returning the result, cancel all workers
	defer cancel()

	// calculate workers count
	if workers <= 0 || workers > len(items) {
		workers = len(items)
	}

	// run workers
	jobs := make(chan int, len(items))
	result := make(chan bool, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(jobs, result, ctx)
	}

	// close the result channel when all workers have done
	go func() {
		wg.Wait()
		close(result)
	}()

	// schedule the jobs: indices to check
	for i := 0; i < len(items); i++ {
		jobs <- i
	}
	close(jobs)

	for range result {
		return false
	}
	return true
}

// AnyAsync returns true if f returns true for any element in items.
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
func AnyAsync[S ~[]T, T any](items S, workers int, f func(el T) bool) bool {
	if len(items) == 0 {
		return false
	}

	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int, result chan<- bool, ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case index, ok := <-jobs:
				if !ok {
					return
				}
				if f(items[index]) {
					result <- true
					return
				}
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// when we're returning the result, cancel all workers
	defer cancel()

	// calculate workers count
	if workers <= 0 || workers > len(items) {
		workers = len(items)
	}

	// run workers
	jobs := make(chan int, len(items))
	result := make(chan bool, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(jobs, result, ctx)
	}

	// close the result channel when all workers have done
	go func() {
		wg.Wait()
		close(result)
	}()

	// schedule the jobs: indices to check
	for i := 0; i < len(items); i++ {
		jobs <- i
	}
	close(jobs)

	for range result {
		return true
	}
	return false
}

// EachAsync calls f for each element in items.
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
func EachAsync[S ~[]T, T any](items S, workers int, f func(el T)) {
	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int) {
		defer wg.Done()
		for index := range jobs {
			f(items[index])
		}
	}

	// calculate workers count
	if workers <= 0 || workers > len(items) {
		workers = len(items)
	}

	// run workers
	jobs := make(chan int, len(items))
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(jobs)
	}

	// add indices into jobs for workers
	for i := 0; i < len(items); i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

// FilterAsync returns a slice containing only items where f returns true
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
//
// The resulting items have the same order as in the input slice.
func FilterAsync[S ~[]T, T any](items S, workers int, f func(el T) bool) S {
	resultMap := make([]bool, len(items))
	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int) {
		for index := range jobs {
			if f(items[index]) {
				resultMap[index] = true
			}
		}
		wg.Done()
	}

	// calculate workers count
	if workers <= 0 || workers > len(items) {
		workers = len(items)
	}

	// run workers
	jobs := make(chan int, len(items))
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(jobs)
	}

	// add indices into jobs for workers
	for i := 0; i < len(items); i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	// return filtered results
	result := make([]T, 0, len(items))
	for i, el := range items {
		if resultMap[i] {
			result = append(result, el)
		}
	}
	return result
}

// MapAsync applies f to all elements in items and returns a slice of the results.
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
//
// The result items have the same order as in the input slice.
func MapAsync[S ~[]T, T any, G any](items S, workers int, f func(el T) G) []G {
	result := make([]G, len(items))
	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int) {
		for index := range jobs {
			result[index] = f(items[index])
		}
		wg.Done()
	}

	// calculate workers count
	if workers <= 0 || workers > len(items) {
		workers = len(items)
	}

	// run workers
	jobs := make(chan int, len(items))
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker(jobs)
	}

	// add indices into jobs for workers
	for i := 0; i < len(items); i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return result
}

// ReduceAsync reduces items to a single value with f.
//
// This is an asynchronous function. It will spawn as many goroutines as you specify
// in the `workers` argument. Set it to zero to spawn a new goroutine for each item.
//
// The function is guaranteed to be called with neighbored items. However, it may be called
// out of order. The results are collected into a new slice which is reduced again, until
// only one item remains. You can think about it as a piramid. On each iteration,
// 2 elements ar taken and merged together until only one remains.
//
// An example for sum:
//
//	1 2 3 4 5
//	 3   7  5
//	  10    5
//	     15
func ReduceAsync[S ~[]T, T any](items S, workers int, f func(left T, right T) T) T {
	if len(items) == 0 {
		var tmp T
		return tmp
	}

	state := make([]T, len(items))
	copy(state, items)
	wg := sync.WaitGroup{}

	worker := func(jobs <-chan int, result chan<- T) {
		for index := range jobs {
			result <- f(state[index], state[index+1])
		}
		wg.Done()
	}

	for len(state) > 1 {
		// calculate workers count
		if workers <= 0 || workers > len(state) {
			workers = len(state)
		}

		// run workers
		jobs := make(chan int, len(state))
		wg.Add(workers)
		result := make(chan T)
		for i := 0; i < workers; i++ {
			go worker(jobs, result)
		}

		go func() {
			wg.Wait()
			close(result)
		}()

		// add indices into jobs for workers
		for i := 0; i < len(state)-1; i += 2 {
			jobs <- i
		}
		close(jobs)

		// collect new state
		newState := make([]T, 0, len(state)/2+len(state)%2)
		for el := range result {
			newState = append(newState, el)
		}
		if len(state)%2 == 1 {
			newState = append(newState, state[len(state)-1])
		}
		// put new state as current state after all
		state = newState
	}

	return state[0]
}
