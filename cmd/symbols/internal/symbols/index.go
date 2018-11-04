package symbols

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"io"
	"io/ioutil"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"golang.org/x/net/trace"
)

func (s *Service) indexedSymbols(ctx context.Context, repo api.RepoName, commitID api.CommitID) (symbols []protocol.Symbol, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "indexedSymbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	key := string(repo) + ":" + string(commitID) + ":v1" // suffix is index format version (vN)

	tr := trace.New("indexedSymbols", string(repo))
	tr.LazyPrintf("commitID: %s", commitID)

	var fetched bool
	defer func() {
		tr.LazyPrintf("fetched=%v symbols=%d", fetched, len(symbols))
		if err != nil {
			tr.LazyPrintf("error: %s", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	f, err := s.cache.Open(ctx, key, func(ctx context.Context) (io.ReadCloser, error) {
		fetched = true

		var err error
		symbols, err = s.parseUncached(ctx, repo, commitID)
		if err != nil {
			return nil, err
		}
		return encodeSymbols(symbols)
	})
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Skip deserializing symbols if we just serialized them in the s.cache.Open call above and still have the
	// slice.
	if symbols == nil {
		var size int64
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}
		tr.LazyPrintf("decode bytes=%d", size)
		symbols, err = decodeSymbols(ctx, f)
		tr.LazyPrintf("decode (done) symbols=%d", len(symbols))
		if err != nil {
			return nil, err
		}
	}

	span.LogFields(otlog.String("event", "result"), otlog.Int("count", len(symbols)))
	return symbols, nil
}

func encodeSymbols(symbols []protocol.Symbol) (io.ReadCloser, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(zw)
	if err := enc.Encode(symbols); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return ioutil.NopCloser(&buf), nil
}

func decodeSymbols(ctx context.Context, r io.Reader) (symbols []protocol.Symbol, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "decodeSymbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	dec := gob.NewDecoder(zr)
	err = dec.Decode(&symbols)
	return symbols, err
}
