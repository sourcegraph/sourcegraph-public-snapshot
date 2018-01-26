// Query and modify DOM storage. (experimental)
package domstorage

import (
	"github.com/neelance/cdp-go/rpc"
)

// Query and modify DOM storage. (experimental)
type Client struct {
	*rpc.Client
}

// DOM Storage identifier. (experimental)

type StorageId struct {
	// Security origin for the storage.
	SecurityOrigin string `json:"securityOrigin"`

	// Whether the storage is local storage (not session storage).
	IsLocalStorage bool `json:"isLocalStorage"`
}

// DOM Storage item. (experimental)

type Item []string

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables storage tracking, storage events will now be delivered to the client.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("DOMStorage.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables storage tracking, prevents storage events from being sent to the client.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("DOMStorage.disable", r.opts, nil)
}

type ClearRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) Clear() *ClearRequest {
	return &ClearRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *ClearRequest) StorageId(v *StorageId) *ClearRequest {
	r.opts["storageId"] = v
	return r
}

func (r *ClearRequest) Do() error {
	return r.client.Call("DOMStorage.clear", r.opts, nil)
}

type GetDOMStorageItemsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) GetDOMStorageItems() *GetDOMStorageItemsRequest {
	return &GetDOMStorageItemsRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetDOMStorageItemsRequest) StorageId(v *StorageId) *GetDOMStorageItemsRequest {
	r.opts["storageId"] = v
	return r
}

type GetDOMStorageItemsResult struct {
	Entries []Item `json:"entries"`
}

func (r *GetDOMStorageItemsRequest) Do() (*GetDOMStorageItemsResult, error) {
	var result GetDOMStorageItemsResult
	err := r.client.Call("DOMStorage.getDOMStorageItems", r.opts, &result)
	return &result, err
}

type SetDOMStorageItemRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) SetDOMStorageItem() *SetDOMStorageItemRequest {
	return &SetDOMStorageItemRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetDOMStorageItemRequest) StorageId(v *StorageId) *SetDOMStorageItemRequest {
	r.opts["storageId"] = v
	return r
}

func (r *SetDOMStorageItemRequest) Key(v string) *SetDOMStorageItemRequest {
	r.opts["key"] = v
	return r
}

func (r *SetDOMStorageItemRequest) Value(v string) *SetDOMStorageItemRequest {
	r.opts["value"] = v
	return r
}

func (r *SetDOMStorageItemRequest) Do() error {
	return r.client.Call("DOMStorage.setDOMStorageItem", r.opts, nil)
}

type RemoveDOMStorageItemRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

func (d *Client) RemoveDOMStorageItem() *RemoveDOMStorageItemRequest {
	return &RemoveDOMStorageItemRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *RemoveDOMStorageItemRequest) StorageId(v *StorageId) *RemoveDOMStorageItemRequest {
	r.opts["storageId"] = v
	return r
}

func (r *RemoveDOMStorageItemRequest) Key(v string) *RemoveDOMStorageItemRequest {
	r.opts["key"] = v
	return r
}

func (r *RemoveDOMStorageItemRequest) Do() error {
	return r.client.Call("DOMStorage.removeDOMStorageItem", r.opts, nil)
}

func init() {
	rpc.EventTypes["DOMStorage.domStorageItemsCleared"] = func() interface{} { return new(DomStorageItemsClearedEvent) }
	rpc.EventTypes["DOMStorage.domStorageItemRemoved"] = func() interface{} { return new(DomStorageItemRemovedEvent) }
	rpc.EventTypes["DOMStorage.domStorageItemAdded"] = func() interface{} { return new(DomStorageItemAddedEvent) }
	rpc.EventTypes["DOMStorage.domStorageItemUpdated"] = func() interface{} { return new(DomStorageItemUpdatedEvent) }
}

type DomStorageItemsClearedEvent struct {
	StorageId *StorageId `json:"storageId"`
}

type DomStorageItemRemovedEvent struct {
	StorageId *StorageId `json:"storageId"`

	Key string `json:"key"`
}

type DomStorageItemAddedEvent struct {
	StorageId *StorageId `json:"storageId"`

	Key string `json:"key"`

	NewValue string `json:"newValue"`
}

type DomStorageItemUpdatedEvent struct {
	StorageId *StorageId `json:"storageId"`

	Key string `json:"key"`

	OldValue string `json:"oldValue"`

	NewValue string `json:"newValue"`
}
