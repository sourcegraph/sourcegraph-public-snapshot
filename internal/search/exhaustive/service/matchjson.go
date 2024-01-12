package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// NewJSONWriter creates a MatchJSONWriter which appends matches to a JSON array
// and uploads them to the object store once the internal buffer size has
// reached 100 MiB or Flush() is called. The object key combines a prefix with
// the shard number, except for the first shard where the shard number is
// excluded.
func NewJSONWriter(ctx context.Context, store uploadstore.Store, prefix string) (*MatchJSONWriter, error) {
	blobUploader := &blobUploader{
		ctx:    ctx,
		store:  store,
		prefix: prefix,
		shard:  1,
	}

	return &MatchJSONWriter{
		w: http.NewJSONArrayBuf(1024*1024*100, blobUploader.write)}, nil
}

type MatchJSONWriter struct {
	w *http.JSONArrayBuf
}

func (m MatchJSONWriter) Flush() error {
	return m.w.Flush()
}

func (m MatchJSONWriter) Write(match result.Match) error {
	eventMatch := search.FromMatch(match, nil, true) // chunk matches enabled

	return m.w.Append(eventMatch)
}

type blobUploader struct {
	ctx    context.Context
	store  uploadstore.Store
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
