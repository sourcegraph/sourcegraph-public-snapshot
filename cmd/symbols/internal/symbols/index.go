package symbols

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"io"
	"io/ioutil"

	"golang.org/x/net/trace"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
)

func (s *Service) indexedSymbols(ctx context.Context, repo api.RepoURI, commitID api.CommitID) (symbols []protocol.Symbol, err error) {
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
		symbols, err = decodeSymbols(f)
		tr.LazyPrintf("decode (done) symbols=%d", len(symbols))
		if err != nil {
			return nil, err
		}
	}
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

func decodeSymbols(r io.Reader) ([]protocol.Symbol, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	dec := gob.NewDecoder(zr)
	var symbols []protocol.Symbol
	err = dec.Decode(&symbols)
	return symbols, err
}
