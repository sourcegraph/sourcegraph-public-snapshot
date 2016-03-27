// Package reactbridge renders React components (written in
// JavaScript) from Go and returns their HTML string.
package reactbridge

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/olebedev/go-duktape.v2"
)

// GlobalFuncs are injected into the global JavaScript context.
var GlobalFuncs = map[string]func(*duktape.Context) int{}

type Bridge struct {
	pool
}

// New initializes a JavaScript bridge with the specified JavaScript
// code.
//
// The bridge is initialized asynchronously (so that program startup
// needn't incur the delay of loading the JS VM). If an error occurs
// during initialization, it is passed to the errCh. If errCh is nil,
// an error will result in a panic.
func New(js string, poolSize int, errCh chan<- error) *Bridge {
	if errCh == nil {
		errCh2 := make(chan error)
		errCh = errCh2
		go func() {
			for err := range errCh2 {
				panic("JavaScript bridge init: " + err.Error())
			}
		}()
	}
	return &Bridge{pool: newDuktapePool(js, poolSize, errCh)}
}

// CallMain calls the function named "main" in the JavaScript VM
// context. It passes 2 arguments to main: the JSON object
// representation of arg, and a callback function. When the callback
// function is called by the "main" function, CallMain returns the
// argument provided to the calback function.
func (b *Bridge) CallMain(ctx context.Context, arg interface{}) (string, error) {
	poolStart := time.Now()
	var vm *vmContext
	select {
	case vm = <-b.get():
	case <-ctx.Done():
		log15.Warn("JavaScript VM pool get timed out", "poolTime", time.Since(poolStart))
		return "", ctx.Err()
	}

	evalStart := time.Now()
	resCh2 := make(chan string)
	errCh := make(chan error)

	go func() {
		resCh, err := vm.Handle(arg)
		if err != nil {
			errCh <- err
			return
		}
		// Forward to the outer channel.
		resCh2 <- (<-resCh)
	}()

	select {
	case res := <-resCh2:
		if evalTime := time.Since(evalStart); evalTime > 750*time.Millisecond {
			log15.Warn("JavaScript eval", "evalTime", evalTime, "poolTime", evalStart.Sub(poolStart))
		}
		b.put(vm)
		return res, nil

	case err := <-errCh:
		log15.Info("JavaScript eval threw error", "error", err, "evalTime", time.Since(evalStart))

		// Drop the VM in case it's corrupted. (If we can guarantee
		// it's not corrupted after an error, we can just use
		// b.put(vm).)
		if err2 := b.drop(vm); err2 != nil {
			log15.Warn("Error releasing JavaScript VM context after error", "error", err, "underlyingError", err2)
		}

		return "", err

	case <-ctx.Done():
		log15.Warn("JavaScript eval timed out", "evalTime", time.Since(evalStart), "poolTime", evalStart.Sub(poolStart))

		// Release JavaScript VM context.
		if err := b.drop(vm); err != nil {
			log15.Warn("Error releasing JavaScript VM context after timeout", "error", err)
			return "", err
		}

		return "", ctx.Err()
	}
}

// Close releases the JavaScript VM contexts managed by the bridge.
func (b *Bridge) Close() error {
	t := time.AfterFunc(5*time.Second, func() {
		log15.Warn("Closing JavaScript bridge is taking a while")
	})
	defer t.Stop()

	return b.dropAll()
}

// newVMContext loads the JavaScript code and creates a new JavaScript
// VM context.
func newVMContext(js string) (*vmContext, error) {
	vm := &vmContext{
		Context: duktape.New(),
		resCh:   make(chan string, 1),
	}

	if err := vm.PushTimers(); err != nil {
		return nil, err
	}

	if err := vm.PevalString(`var console = {log:print,warn:print,error:print,info:print,count:function(){},time:function(){},timeEnd:function(){}};`); err != nil {
		return nil, err
	}

	goCallback := func(dctx *duktape.Context) int {
		result := dctx.SafeToString(-1)
		vm.resCh <- string(result)
		return 0
	}
	if _, err := vm.PushGlobalGoFunction("__goCallback__", goCallback); err != nil {
		return nil, err
	}
	for name, fn := range GlobalFuncs {
		if _, err := vm.PushGlobalGoFunction(name, fn); err != nil {
			return nil, err
		}
	}

	if err := vm.PevalLstring(js, len(js)); err != nil {
		log15.Error("Error evaluating JavaScript", "error", err)
		return nil, err
	}

	vm.PopN(vm.GetTop())
	return vm, nil
}

// vmContext wraps duktape.Context
type vmContext struct {
	*duktape.Context
	resCh chan string
}

// Handle handles http requests
func (vm *vmContext) Handle(arg interface{}) (<-chan string, error) {
	b, err := json.Marshal(arg)
	if err != nil {
		return nil, err
	}
	js := `main(` + string(b) + `, __goCallback__);`
	vm.Context.Lock()
	defer vm.Context.Unlock()
	if err := vm.PevalLstring(js, len(js)); err != nil {
		if derr, ok := err.(*duktape.Error); ok {
			return nil, &Error{err: *derr}
		}
		return nil, err
	}
	return vm.resCh, nil
}

// DestroyHeap destroys the context's heap
func (vm *vmContext) DestroyHeap() {
	close(vm.resCh)
	vm.Context.DestroyHeap()
}

// Error wraps duktape.Error to provide more information in its Error
// method.
type Error struct {
	err duktape.Error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s [%s:%d] %s", e.err.Type, e.err.Message, e.err.FileName, e.err.LineNumber, e.err.Stack)
}

type pool interface {
	get() <-chan *vmContext
	put(*vmContext)
	drop(*vmContext) error
	dropAll() error
}

func newDuktapePool(js string, size int, errCh chan<- error) *duktapePool {
	pool := &duktapePool{
		js:   js,
		ch:   make(chan *vmContext, size),
		size: size,
	}

	var wg sync.WaitGroup
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vm, err := newVMContext(js)
			if err != nil {
				errCh <- err
				return
			}
			pool.ch <- vm
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	return pool
}

type duktapePool struct {
	ch   chan *vmContext
	js   string
	size int
}

func (o *duktapePool) get() <-chan *vmContext {
	return o.ch
}

func (o *duktapePool) put(ctx *vmContext) {
	// Drop any future async calls.
	ctx.Lock()
	ctx.FlushTimers()
	ctx.Unlock()
	o.ch <- ctx
}

func (o *duktapePool) drop(ctx *vmContext) error {
	ctx.Lock()
	ctx.FlushTimers()
	ctx.Gc(0)
	ctx.DestroyHeap()
	ctx = nil
	vm, err := newVMContext(o.js)
	if err != nil {
		return err
	}
	o.ch <- vm
	return nil
}

func (o *duktapePool) dropAll() error {
	var anyErr error
	for i := 0; i < o.size; i++ {
		vm := <-o.ch
		if err := o.drop(vm); err != nil {
			anyErr = err
		}
	}
	return anyErr
}
