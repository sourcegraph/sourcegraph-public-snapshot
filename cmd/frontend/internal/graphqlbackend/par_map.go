package graphqlbackend

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// f: int -> B, values: []A, results []B. f should retrieve a value of type A
// from values, and return a value of type B.
func ParMap(f, values interface{}) (interface{}, error) {
	vals := reflect.ValueOf(values)
	if vals.Kind() != reflect.Slice {
		panic("arguments to parmap must be slice")
	}
	l := vals.Len()
	fV := reflect.ValueOf(f)
	fT := reflect.TypeOf(f)
	if fT.NumOut() != 2 || fT.NumIn() != 1 {
		panic("f must take 1 arg and return 1 value")
	}
	ret := reflect.MakeSlice(reflect.SliceOf(fT.Out(0)), l, l)
	numWorkers := 4
	if numWorkers > l {
		numWorkers = l
	}
	var n uint64 = ^uint64(0)
	wg := sync.WaitGroup{}
	wg.Add(numWorkers)
	var atomicErr atomic.Value
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				index := int(atomic.AddUint64(&n, 1))
				if index >= l {
					break
				}
				a := vals.Index(index)
				v := fV.Call([]reflect.Value{a})
				if !v[1].IsNil() {
					atomicErr.Store(v[1].Interface())
					atomic.SwapUint64(&n, uint64(l))
				}
				ret.Index(index).Set(v[0])
			}
			wg.Done()
		}()
	}
	wg.Wait()
	err := atomicErr.Load()
	if err != nil {
		return nil, err.(error)
	}
	return ret.Interface(), nil
}
