package rwvfs

import "golang.org/x/tools/godoc/vfs"

// FetcherOpener is implemented by FileSystems that support a mode of
// operation where OpenFetcher returns a lazily loaded file. This is
// useful for VFS implementations that are backed by slow (e.g.,
// network) data sources, so you can explicit fetch byte ranges at a
// higher level than Go I/O buffering.
type FetcherOpener interface {
	OpenFetcher(name string) (vfs.ReadSeekCloser, error)
}

// Fetcher is implemented by files that require explicit fetching to
// buffer data from remote (network-like) underlying sources.
type Fetcher interface {
	// Fetch fetches the specified byte range (start inclusive, end
	// exclusive) from a remote (network-like) underlying source. Only
	// these bytes ranges are available to be read.
	Fetch(start, end int64) error
}
