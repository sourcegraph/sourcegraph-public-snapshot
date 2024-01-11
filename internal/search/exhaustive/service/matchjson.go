package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// NewJSONWriter returns a MatchWriter that serializes matches to JSON and
// writes them to the store. Files are rotated every 100MB. The key is the
// prefix + shard number. The shard number is omitted for the first shard.
func NewJSONWriter(ctx context.Context, store uploadstore.Store, prefix string) (*MatchJSONWriter, error) {
	blobUploader := &blobUploader{
		ctx:    ctx,
		store:  store,
		prefix: prefix,
		shard:  1,
	}

	bufferedWriter := &bufferedWriter{
		maxSizeBytes: 100_000_000, // 100MB
		w:            blobUploader,
	}

	return &MatchJSONWriter{
		w: bufferedWriter,
	}, nil
}

type MatchJSONWriter struct {
	w io.WriteCloser
}

func (m MatchJSONWriter) Close() error {
	return m.w.Close()
}

func (m MatchJSONWriter) Write(match result.Match) error {
	eventMatch := search.FromMatch(match, nil, true) // chunk matches enabled

	// TODO (stefan) check that this is how the Stream API does it.
	b, err := json.Marshal(eventMatch)
	if err != nil {
		return err
	}

	_, err = m.w.Write(b)
	if err != nil {
		return err
	}

	_, err = m.w.Write([]byte("\n"))
	return err
}

type blobUploader struct {
	ctx    context.Context
	store  uploadstore.Store
	prefix string
	shard  int
}

func (b *blobUploader) Write(p []byte) (int, error) {
	key := ""
	if b.shard == 1 {
		key = fmt.Sprintf("%s", b.prefix)
	} else {
		key = fmt.Sprintf("%s-%d", b.prefix, b.shard)
	}
	n64, err := b.store.Upload(b.ctx, key, bytes.NewBuffer(p))
	if err != nil {
		return int(n64), err
	}

	b.shard += 1

	return int(n64), nil
}

// bufferedWriter is a writer that will write to the underlying writer once the
// total number of bytes written exceeds maxSizeBytes.
type bufferedWriter struct {
	maxSizeBytes int64
	buf          bytes.Buffer
	w            io.Writer
}

func (w *bufferedWriter) Close() error {
	return w.flush()
}

func (w *bufferedWriter) flush() error {
	_, err := w.w.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

func (w *bufferedWriter) Write(b []byte) (int, error) {
	if int64(w.buf.Len()) >= w.maxSizeBytes {
		err := w.flush()
		if err != nil {
			return 0, err
		}
	}

	return w.buf.Write(b)
}
