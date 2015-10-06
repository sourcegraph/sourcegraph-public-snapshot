package vfsutil

import "golang.org/x/tools/godoc/vfs"

type readRet struct {
	Bytes []byte
	Error error
}

const parSize = 4

// ConcurrentRead will call vfs.ReadFile in concurrent batches and return the
// bytes over the channel in the order the paths are specified. Users must
// call doneSignal.Done() to ensure cleanup of all goroutines.
//
// 	readCh, done := ConcurrentRead(opener, paths)
// 	defer done.Done()
// 	for ret := range readCh {
// 		if ret.Error != nil {
// 			return ret.Error
// 		}
// 		// do something with ret.Bytes
// 	}
//
// Up to parSize paths will be read ahead of time. If you stop reading early,
// this interface will cancel any further concurrent reads
func ConcurrentRead(fs vfs.Opener, paths []string) (chan readRet, doneSignal) {
	size := parSize
	if len(paths) < size {
		size = len(paths)
	}
	done := doneSignal{make(chan struct{}, size+1), size + 1}
	out := make(chan readRet)
	res := make([]readRet, size)
	sem := make([]chan struct{}, size)

	read := func(path string, n int) {
		b, err := vfs.ReadFile(fs, path)
		res[n].Bytes = b
		res[n].Error = err
		select {
		case sem[n] <- struct{}{}:
		case <-done.done:
		}
	}
	for n := 0; n < size; n++ {
		sem[n] = make(chan struct{})
		go read(paths[n], n)
	}
	go func() {
		for i := 0; i < len(paths); i++ {
			n := i % size
			<-sem[n]
			select {
			case out <- res[n]:
			case <-done.done:
				// This will cause us to break out of the for
				// loop
				i = len(paths)
			}
			next := i + size
			if next < len(paths) {
				go read(paths[next], n)
			}
		}
		close(out)
	}()
	return out, done
}

type doneSignal struct {
	done chan struct{}
	size int
}

func (d *doneSignal) Done() {
	for i := 0; i < d.size; i++ {
		d.done <- struct{}{}
	}
}
