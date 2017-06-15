// (experimental)
package cachestorage

import (
	"github.com/neelance/cdp-go/rpc"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Unique identifier of the Cache object.

type CacheId string

// Data entry.

type DataEntry struct {
	// Request url spec.
	Request string `json:"request"`

	// Response status text.
	Response string `json:"response"`

	// Number of seconds since epoch.
	ResponseTime float64 `json:"responseTime"`
}

// Cache identifier.

type Cache struct {
	// An opaque unique id of the cache.
	CacheId CacheId `json:"cacheId"`

	// Security origin of the cache.
	SecurityOrigin string `json:"securityOrigin"`

	// The name of the cache.
	CacheName string `json:"cacheName"`
}

type RequestCacheNamesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests cache names.
func (d *Client) RequestCacheNames() *RequestCacheNamesRequest {
	return &RequestCacheNamesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *RequestCacheNamesRequest) SecurityOrigin(v string) *RequestCacheNamesRequest {
	r.opts["securityOrigin"] = v
	return r
}

type RequestCacheNamesResult struct {
	// Caches for the security origin.
	Caches []*Cache `json:"caches"`
}

func (r *RequestCacheNamesRequest) Do() (*RequestCacheNamesResult, error) {
	var result RequestCacheNamesResult
	err := r.client.Call("CacheStorage.requestCacheNames", r.opts, &result)
	return &result, err
}

type RequestEntriesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests data from cache.
func (d *Client) RequestEntries() *RequestEntriesRequest {
	return &RequestEntriesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// ID of cache to get entries from.
func (r *RequestEntriesRequest) CacheId(v CacheId) *RequestEntriesRequest {
	r.opts["cacheId"] = v
	return r
}

// Number of records to skip.
func (r *RequestEntriesRequest) SkipCount(v int) *RequestEntriesRequest {
	r.opts["skipCount"] = v
	return r
}

// Number of records to fetch.
func (r *RequestEntriesRequest) PageSize(v int) *RequestEntriesRequest {
	r.opts["pageSize"] = v
	return r
}

type RequestEntriesResult struct {
	// Array of object store data entries.
	CacheDataEntries []*DataEntry `json:"cacheDataEntries"`

	// If true, there are more entries to fetch in the given range.
	HasMore bool `json:"hasMore"`
}

func (r *RequestEntriesRequest) Do() (*RequestEntriesResult, error) {
	var result RequestEntriesResult
	err := r.client.Call("CacheStorage.requestEntries", r.opts, &result)
	return &result, err
}

type DeleteCacheRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Deletes a cache.
func (d *Client) DeleteCache() *DeleteCacheRequest {
	return &DeleteCacheRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of cache for deletion.
func (r *DeleteCacheRequest) CacheId(v CacheId) *DeleteCacheRequest {
	r.opts["cacheId"] = v
	return r
}

func (r *DeleteCacheRequest) Do() error {
	return r.client.Call("CacheStorage.deleteCache", r.opts, nil)
}

type DeleteEntryRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Deletes a cache entry.
func (d *Client) DeleteEntry() *DeleteEntryRequest {
	return &DeleteEntryRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of cache where the entry will be deleted.
func (r *DeleteEntryRequest) CacheId(v CacheId) *DeleteEntryRequest {
	r.opts["cacheId"] = v
	return r
}

// URL spec of the request.
func (r *DeleteEntryRequest) Request(v string) *DeleteEntryRequest {
	r.opts["request"] = v
	return r
}

func (r *DeleteEntryRequest) Do() error {
	return r.client.Call("CacheStorage.deleteEntry", r.opts, nil)
}

func init() {
}
