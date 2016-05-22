package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

// clientCacheInfo reads the cache information supplied by the client in the
// HTTP request.
func clientCacheInfo(r *http.Request) (ifModSince time.Time, noCache bool, err error) {
	if timeStr := r.Header.Get("if-modified-since"); timeStr != "" {
		ifModSince, err = http.ParseTime(timeStr)
	}
	noCache = strings.Contains(r.Header.Get("cache-control"), "no-cache")
	return ifModSince, noCache, err
}

var (
	// defaultCacheMaxAge is the default "Cache-Control: max-age=N" value to
	// send as an HTTP response header. During development, this should be set
	// to 0 to allow more accurate performance measurement.
	defaultCacheMaxAge = conf.GetenvDurationOrDefault("SG_API_DEFAULT_CACHE_MAX_AGE", "0s")
)

// writeCacheHeaders writes HTTP cache-related response headers.
//
// If lastMod is zero ((time.Time).IsZero() returns true on it), then only a
// max-age Cache-Control header is written.
//
// If the client's If-Modified-Since date (clientMod) is equal to or newer than
// the server's Last-Modified (lastMod), it writes a HTTP 304 Not Modified
// response.
//
// If clientCached is true, the request handler should immediately return. (This is used
// when writeCacheHeaders returns an error or writes an HTTP 304 Not Modified.
// In these cases, no further handling should occur.)
//
// Thus handlers should call writeCacheHeaders as follows:
//
// 	if clientCached, err := writeCacheHeaders(w, r, lastMod, maxAge); clientCached || err != nil {
//	  return err
//	}
func writeCacheHeaders(w http.ResponseWriter, r *http.Request, lastMod time.Time, maxAge time.Duration) (clientCached bool, err error) {
	w.Header().Set("cache-control", fmt.Sprintf("private, max-age=%d", maxAge/time.Second))

	if lastMod.IsZero() {
		return false, nil
	}

	clientMod, noCache, err := clientCacheInfo(r)
	if err != nil {
		return true, err
	}

	lastMod = lastMod.Round(time.Second)
	w.Header().Set("last-modified", lastMod.Format(http.TimeFormat))
	if !noCache && (lastMod.Equal(clientMod) || lastMod.Before(clientMod)) {
		w.WriteHeader(http.StatusNotModified)
		return true, nil
	}

	return false, nil
}

// getLastModForRepoRevs gets the most recent last build end time for
// any repo/commit pair in repoRevs (which is a list of strings of the
// form "uri[@rev]", like "repohost.com/foo@master" or just
// "repohost.com/foo").
func getLastModForRepoRevs(r *http.Request, repoRevs []string) (time.Time, error) {
	if len(repoRevs) == 1 {
		repoURI, commitID := sourcegraph.ParseRepoAndCommitID(repoRevs[0])
		if commitID != "" {
			// Only stats could have changed since the build completed, so set a
			// long max-age.
			//
			// TODO(sqs): perf can be improved by adding cache headers in
			// the case where multiple repo URIs are specified (currently
			// this logic is only if 1 repo is specified).
			lastMod, err := getRepoLastBuildTime(r, sourcegraph.RepoSpec{URI: repoURI}, commitID)
			if err != nil {
				return time.Time{}, err
			}
			return lastMod, nil
		}
	}
	return time.Time{}, nil
}
