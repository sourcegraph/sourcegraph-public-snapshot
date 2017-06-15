// (experimental)
package indexeddb

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/runtime"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Database with an array of object stores.

type DatabaseWithObjectStores struct {
	// Database name.
	Name string `json:"name"`

	// Database version.
	Version int `json:"version"`

	// Object stores in this database.
	ObjectStores []*ObjectStore `json:"objectStores"`
}

// Object store.

type ObjectStore struct {
	// Object store name.
	Name string `json:"name"`

	// Object store key path.
	KeyPath *KeyPath `json:"keyPath"`

	// If true, object store has auto increment flag set.
	AutoIncrement bool `json:"autoIncrement"`

	// Indexes in this object store.
	Indexes []*ObjectStoreIndex `json:"indexes"`
}

// Object store index.

type ObjectStoreIndex struct {
	// Index name.
	Name string `json:"name"`

	// Index key path.
	KeyPath *KeyPath `json:"keyPath"`

	// If true, index is unique.
	Unique bool `json:"unique"`

	// If true, index allows multiple entries for a key.
	MultiEntry bool `json:"multiEntry"`
}

// Key.

type Key struct {
	// Key type.
	Type string `json:"type"`

	// Number value. (optional)
	Number float64 `json:"number,omitempty"`

	// String value. (optional)
	String string `json:"string,omitempty"`

	// Date value. (optional)
	Date float64 `json:"date,omitempty"`

	// Array value. (optional)
	Array []*Key `json:"array,omitempty"`
}

// Key range.

type KeyRange struct {
	// Lower bound. (optional)
	Lower *Key `json:"lower,omitempty"`

	// Upper bound. (optional)
	Upper *Key `json:"upper,omitempty"`

	// If true lower bound is open.
	LowerOpen bool `json:"lowerOpen"`

	// If true upper bound is open.
	UpperOpen bool `json:"upperOpen"`
}

// Data entry.

type DataEntry struct {
	// Key object.
	Key *runtime.RemoteObject `json:"key"`

	// Primary key object.
	PrimaryKey *runtime.RemoteObject `json:"primaryKey"`

	// Value object.
	Value *runtime.RemoteObject `json:"value"`
}

// Key path.

type KeyPath struct {
	// Key path type.
	Type string `json:"type"`

	// String value. (optional)
	String string `json:"string,omitempty"`

	// Array value. (optional)
	Array []string `json:"array,omitempty"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables events from backend.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("IndexedDB.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables events from backend.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("IndexedDB.disable", r.opts, nil)
}

type RequestDatabaseNamesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests database names for given security origin.
func (d *Client) RequestDatabaseNames() *RequestDatabaseNamesRequest {
	return &RequestDatabaseNamesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *RequestDatabaseNamesRequest) SecurityOrigin(v string) *RequestDatabaseNamesRequest {
	r.opts["securityOrigin"] = v
	return r
}

type RequestDatabaseNamesResult struct {
	// Database names for origin.
	DatabaseNames []string `json:"databaseNames"`
}

func (r *RequestDatabaseNamesRequest) Do() (*RequestDatabaseNamesResult, error) {
	var result RequestDatabaseNamesResult
	err := r.client.Call("IndexedDB.requestDatabaseNames", r.opts, &result)
	return &result, err
}

type RequestDatabaseRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests database with given name in given frame.
func (d *Client) RequestDatabase() *RequestDatabaseRequest {
	return &RequestDatabaseRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *RequestDatabaseRequest) SecurityOrigin(v string) *RequestDatabaseRequest {
	r.opts["securityOrigin"] = v
	return r
}

// Database name.
func (r *RequestDatabaseRequest) DatabaseName(v string) *RequestDatabaseRequest {
	r.opts["databaseName"] = v
	return r
}

type RequestDatabaseResult struct {
	// Database with an array of object stores.
	DatabaseWithObjectStores *DatabaseWithObjectStores `json:"databaseWithObjectStores"`
}

func (r *RequestDatabaseRequest) Do() (*RequestDatabaseResult, error) {
	var result RequestDatabaseResult
	err := r.client.Call("IndexedDB.requestDatabase", r.opts, &result)
	return &result, err
}

type RequestDataRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests data from object store or index.
func (d *Client) RequestData() *RequestDataRequest {
	return &RequestDataRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *RequestDataRequest) SecurityOrigin(v string) *RequestDataRequest {
	r.opts["securityOrigin"] = v
	return r
}

// Database name.
func (r *RequestDataRequest) DatabaseName(v string) *RequestDataRequest {
	r.opts["databaseName"] = v
	return r
}

// Object store name.
func (r *RequestDataRequest) ObjectStoreName(v string) *RequestDataRequest {
	r.opts["objectStoreName"] = v
	return r
}

// Index name, empty string for object store data requests.
func (r *RequestDataRequest) IndexName(v string) *RequestDataRequest {
	r.opts["indexName"] = v
	return r
}

// Number of records to skip.
func (r *RequestDataRequest) SkipCount(v int) *RequestDataRequest {
	r.opts["skipCount"] = v
	return r
}

// Number of records to fetch.
func (r *RequestDataRequest) PageSize(v int) *RequestDataRequest {
	r.opts["pageSize"] = v
	return r
}

// Key range. (optional)
func (r *RequestDataRequest) KeyRange(v *KeyRange) *RequestDataRequest {
	r.opts["keyRange"] = v
	return r
}

type RequestDataResult struct {
	// Array of object store data entries.
	ObjectStoreDataEntries []*DataEntry `json:"objectStoreDataEntries"`

	// If true, there are more entries to fetch in the given range.
	HasMore bool `json:"hasMore"`
}

func (r *RequestDataRequest) Do() (*RequestDataResult, error) {
	var result RequestDataResult
	err := r.client.Call("IndexedDB.requestData", r.opts, &result)
	return &result, err
}

type ClearObjectStoreRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Clears all entries from an object store.
func (d *Client) ClearObjectStore() *ClearObjectStoreRequest {
	return &ClearObjectStoreRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *ClearObjectStoreRequest) SecurityOrigin(v string) *ClearObjectStoreRequest {
	r.opts["securityOrigin"] = v
	return r
}

// Database name.
func (r *ClearObjectStoreRequest) DatabaseName(v string) *ClearObjectStoreRequest {
	r.opts["databaseName"] = v
	return r
}

// Object store name.
func (r *ClearObjectStoreRequest) ObjectStoreName(v string) *ClearObjectStoreRequest {
	r.opts["objectStoreName"] = v
	return r
}

func (r *ClearObjectStoreRequest) Do() error {
	return r.client.Call("IndexedDB.clearObjectStore", r.opts, nil)
}

type DeleteDatabaseRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Deletes a database.
func (d *Client) DeleteDatabase() *DeleteDatabaseRequest {
	return &DeleteDatabaseRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Security origin.
func (r *DeleteDatabaseRequest) SecurityOrigin(v string) *DeleteDatabaseRequest {
	r.opts["securityOrigin"] = v
	return r
}

// Database name.
func (r *DeleteDatabaseRequest) DatabaseName(v string) *DeleteDatabaseRequest {
	r.opts["databaseName"] = v
	return r
}

func (r *DeleteDatabaseRequest) Do() error {
	return r.client.Call("IndexedDB.deleteDatabase", r.opts, nil)
}

func init() {
}
