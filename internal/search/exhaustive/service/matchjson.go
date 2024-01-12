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

// NewJSONWriter returns a MatchJSONWriter that serializes matches to a JSON
// array and writes them to the store. Matches are uploaded as blobs with a max
// size of 100MB. The object key is the prefix + shard number. For the first
// shard, the shard number is omitted.
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

func (m MatchJSONWriter) Close() error {
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
