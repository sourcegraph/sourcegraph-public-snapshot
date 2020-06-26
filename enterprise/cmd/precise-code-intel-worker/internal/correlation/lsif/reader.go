package lsif

import (
	"context"
	"io"
	"runtime"

	simdjson "github.com/minio/simdjson-go"
)

// ChannelBufferSize is the size of the parsed elements channel.
const ChannelBufferSize = 1024

// NumUnmarshalGoRoutines is the number of goroutines that can read parsed streams concurrently.
var ReadConcurrency = runtime.NumCPU() * 2

// reuse is a channel used to give unused ParsedJson objects back to simdjson for reuse.
var reuse = make(chan *simdjson.ParsedJson, ChannelBufferSize)

// Pair holds either a parsed element or an error that occurred during parsing.
type Pair struct {
	Element Element
	Err     error
}

// TODO - document
func Read(ctx context.Context, r io.Reader) <-chan Pair {
	// TODO - ensure reader is closed and stream drains on error
	stream := make(chan simdjson.Stream, ChannelBufferSize)
	simdjson.ParseNDStream(r, stream, reuse)

	pairCh := make(chan Pair, ChannelBufferSize)

	go func() {
		defer close(pairCh)

		for elem := range stream {
			if elem.Error != nil {
				if elem.Error == io.EOF {
					return
				}

				select {
				case pairCh <- Pair{Err: elem.Error}:
				case <-ctx.Done():
				}
				return
			}

			iter := elem.Value.Iter()

			for iter.Advance() != simdjson.TypeNone {
				var temp simdjson.Iter
				_, iter, err := iter.Root(&temp)
				if err != nil {
					select {
					case pairCh <- Pair{Err: err}:
					case <-ctx.Done():
					}
					return
				}

				element, err := unmarshalElement(iter)
				if err != nil {
					select {
					case pairCh <- Pair{Err: err}:
					case <-ctx.Done():
					}
					return
				}

				select {
				case pairCh <- Pair{Element: element}:
				case <-ctx.Done():
					return
				}
			}

			select {
			case reuse <- elem.Value:
			default:
			}
		}
	}()

	return pairCh

	// Old:
	// return readQueue(ctx, readStream(ctx, stream))
}

// // TODO - document
// func readStream(ctx context.Context, stream <-chan simdjson.Stream) <-chan chan Pair {
// 	queue := make(chan chan Pair, ReadConcurrency)

// 	go func() {
// 		defer close(queue)

// 		for elem := range stream {
// 			if elem.Error != nil {
// 				if elem.Error != io.EOF {
// 					ch := make(chan Pair, 1)
// 					ch <- Pair{Err: elem.Error}

// 					select {
// 					case queue <- ch:
// 					case <-ctx.Done():
// 					}
// 				}

// 				return
// 			}

// 			iter := elem.Value.Iter()

// 			ch := make(chan Pair, ChannelBufferSize)
// 			select {
// 			case queue <- ch:
// 			case <-ctx.Done():
// 				return
// 			}

// 			/* go */
// 			func(elem simdjson.Stream) {
// 				defer close(ch)

// 				for iter.Advance() != simdjson.TypeNone {
// 					var temp simdjson.Iter
// 					_, iter, err := iter.Root(&temp)
// 					if err != nil {
// 						select {
// 						case ch <- Pair{Err: err}:
// 						case <-ctx.Done():
// 						}

// 						return
// 					}

// 					element, err := unmarshalElement(iter)
// 					if err != nil {
// 						select {
// 						case ch <- Pair{Err: err}:
// 						case <-ctx.Done():
// 						}

// 						return
// 					}

// 					select {
// 					case ch <- Pair{Element: element}:
// 					case <-ctx.Done():
// 						return
// 					}
// 				}

// 				select {
// 				case reuse <- elem.Value:
// 				default:
// 				}
// 			}(elem)
// 		}
// 	}()

// 	return queue
// }

// // TODO - document
// func readQueue(ctx context.Context, queue <-chan chan Pair) chan Pair {
// 	pairCh := make(chan Pair, ChannelBufferSize)

// 	go func() {
// 		defer close(pairCh)

// 		for {
// 			select {
// 			case ch, ok := <-queue:
// 				if !ok {
// 					return
// 				}

// 			streamLoop:
// 				for {
// 					select {
// 					case pair, ok := <-ch:
// 						if !ok {
// 							break streamLoop
// 						}

// 						select {
// 						case pairCh <- pair:
// 						case <-ctx.Done():
// 							return
// 						}

// 					case <-ctx.Done():
// 						return
// 					}
// 				}

// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()

// 	return pairCh
// }
