package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// NewJSONWriter creates a MatchJSONWriter which appends matches to a JSON array
// and uploads them to the object store once the internal buffer size has
// reached 100 MiB or Flush() is called. The object key combines a prefix with
// the shard number, except for the first shard where the shard number is
// omitted.
func NewJSONWriter(ctx context.Context, store object.Storage, prefix string) (*MatchJSONWriter, error) {
	blobUploader := &blobUploader{
		ctx:    ctx,
		store:  store,
		prefix: prefix,
		shard:  1,
	}

	return &MatchJSONWriter{
		w: newBufferedWriter(1024*1024*100, blobUploader.write)}, nil
}

type MatchJSONWriter struct {
	w *bufferedWriter
}

func (m MatchJSONWriter) Flush() error {
	return m.w.Flush()
}

func (m MatchJSONWriter) Write(match result.Match) error {
	eventMatch := search.FromMatch(match, nil, search.FromMatchOptions{
		ChunkMatches:         true,
		MaxContentLineLength: -1, // do not truncate content
	})

	return m.w.Append(eventMatch)
}

type blobUploader struct {
	ctx    context.Context
	store  object.Storage
	prefix string
	shard  int
}

func (b *blobUploader) write(p []byte) error {
	key := ""
	if b.shard == 1 {
		key = b.prefix
	} else {
		key = fmt.Sprintf("%s-%d", b.prefix, b.shard)
	}

	_, err := b.store.Upload(b.ctx, key, bytes.NewBuffer(p))
	if err != nil {
		return err
	}

	b.shard += 1

	return nil
}

type bufferedWriter struct {
	flushSize int
	buf       bytes.Buffer
	write     func([]byte) error
}

func newBufferedWriter(flushSize int, write func([]byte) error) *bufferedWriter {
	b := &bufferedWriter{
		flushSize: flushSize,
		write:     write,
	}
	// Grow the buffer to reduce the number of small allocations caused by
	// repeatedly growing the buffer. We expect most queries to return a fraction of
	// flushSize per revision.
	b.buf.Grow(min(flushSize, 1024*1024)) // 1 MiB
	return b
}

// Append marshals v and adds it to the buffer. If the size of the buffer
// exceeds flushSize the buffer is written out.
func (j *bufferedWriter) Append(v any) error {
	oldLen := j.buf.Len()

	enc := json.NewEncoder(&j.buf)
	if err := enc.Encode(v); err != nil {
		// Reset the buffer to where it was before failing to marshal
		j.buf.Truncate(oldLen)
		return err
	}

	if j.buf.Len() >= j.flushSize {
		return j.Flush()
	}

	return nil
}

// Flush writes and resets the buffer if there is data to write.
func (j *bufferedWriter) Flush() error {
	if j.buf.Len() == 0 {
		return nil
	}

	buf := j.buf.Bytes()
	j.buf.Reset()

	return j.write(buf)
}

func (j *bufferedWriter) Len() int {
	return j.buf.Len()
}
