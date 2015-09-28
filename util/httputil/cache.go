package httputil

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/peterbourgon/diskv"
	"github.com/sourcegraph/httpcache"
	"github.com/sourcegraph/httpcache/diskcache"
	"github.com/sourcegraph/loggedcache"
	"sourcegraph.com/sourcegraph/multicache"
	"sourcegraph.com/sourcegraph/s3cache"
	"src.sourcegraph.com/sourcegraph/util/fileutil"
)

// httpCacheDir is the directory used for caching HTTP responses. It can be reused
// executions (it is not necessary to create a new random temp dir upon
// startup).
var httpCacheDir = filepath.Join(fileutil.TempDir(), "sourcegraph-http-cache")

var (
	LocalCache = diskcache.NewWithDiskv(diskv.New(diskv.Options{
		BasePath: httpCacheDir,
	}))
	LocalAndRemoteCache = loggedcache.Async{
		Underlying: multicache.NewFallback(multicaches()...),

		// // Uncomment to log cache timings:
		// Log:        log.New(os.Stderr, "CACHE: ", 0),
		// Time:       func(operation string, t time.Duration) {},
	}
)

func multicaches() []multicache.Underlying {
	cs := []multicache.Underlying{LocalCache}
	if bucketURL := os.Getenv("SG_S3_HTTP_CACHE_BUCKET"); bucketURL != "" {
		s3Cache := s3cache.New(bucketURL)
		s3Cache.Gzip = true
		cs = append(cs, s3Cache)
	}
	return cs
}

func init() {
	err := os.Mkdir(httpCacheDir, 0700)
	if err != nil && !os.IsExist(err) {
		log.Panicf("Mkdir(%s) failed: %s", httpCacheDir, err)
	}

	// Only wait to write to and delete from the disk cache (the 1st cache).
	LocalAndRemoteCache.Underlying.(*multicache.Fallback).WaitNSets = 1
	LocalAndRemoteCache.Underlying.(*multicache.Fallback).WaitNDeletes = 1
}

var (
	CachingTransport      = &httpcache.Transport{Cache: &LocalAndRemoteCache}
	CachingHTTPClient     = &http.Client{Transport: CachingTransport}
	LocalCachingTransport = &httpcache.Transport{
		Cache: LocalCache,
	}
)
