// This domain facilitates obtaining document snapshots with DOM, layout, and style information. (experimental)
package domsnapshot

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/css"
	"github.com/neelance/cdp-go/protocol/dom"
)

// This domain facilitates obtaining document snapshots with DOM, layout, and style information. (experimental)
type Client struct {
	*rpc.Client
}

// A Node in the DOM tree.

type DOMNode struct {
	// <code>Node</code>'s nodeType.
	NodeType int `json:"nodeType"`

	// <code>Node</code>'s nodeName.
	NodeName string `json:"nodeName"`

	// <code>Node</code>'s nodeValue.
	NodeValue string `json:"nodeValue"`

	// <code>Node</code>'s id, corresponds to DOM.Node.backendNodeId.
	BackendNodeId dom.BackendNodeId `json:"backendNodeId"`

	// The indexes of the node's child nodes in the <code>domNodes</code> array returned by <code>getSnapshot</code>, if any. (optional)
	ChildNodeIndexes []int `json:"childNodeIndexes,omitempty"`

	// Attributes of an <code>Element</code> node. (optional)
	Attributes []*NameValue `json:"attributes,omitempty"`

	// Indexes of pseudo elements associated with this node in the <code>domNodes</code> array returned by <code>getSnapshot</code>, if any. (optional)
	PseudoElementIndexes []int `json:"pseudoElementIndexes,omitempty"`

	// The index of the node's related layout tree node in the <code>layoutTreeNodes</code> array returned by <code>getSnapshot</code>, if any. (optional)
	LayoutNodeIndex int `json:"layoutNodeIndex,omitempty"`

	// Document URL that <code>Document</code> or <code>FrameOwner</code> node points to. (optional)
	DocumentURL string `json:"documentURL,omitempty"`

	// Base URL that <code>Document</code> or <code>FrameOwner</code> node uses for URL completion. (optional)
	BaseURL string `json:"baseURL,omitempty"`

	// <code>DocumentType</code> node's publicId. (optional)
	PublicId string `json:"publicId,omitempty"`

	// <code>DocumentType</code> node's systemId. (optional)
	SystemId string `json:"systemId,omitempty"`

	// Frame ID for frame owner elements. (optional)
	FrameId string `json:"frameId,omitempty"`

	// The index of a frame owner element's content document in the <code>domNodes</code> array returned by <code>getSnapshot</code>, if any. (optional)
	ContentDocumentIndex int `json:"contentDocumentIndex,omitempty"`

	// Index of the imported document's node of a link element in the <code>domNodes</code> array returned by <code>getSnapshot</code>, if any. (optional)
	ImportedDocumentIndex int `json:"importedDocumentIndex,omitempty"`

	// Index of the content node of a template element in the <code>domNodes</code> array returned by <code>getSnapshot</code>. (optional)
	TemplateContentIndex int `json:"templateContentIndex,omitempty"`

	// Type of a pseudo element node. (optional)
	PseudoType dom.PseudoType `json:"pseudoType,omitempty"`
}

// Details of an element in the DOM tree with a LayoutObject.

type LayoutTreeNode struct {
	// The index of the related DOM node in the <code>domNodes</code> array returned by <code>getSnapshot</code>.
	DomNodeIndex int `json:"domNodeIndex"`

	// The absolute position bounding box.
	BoundingBox *dom.Rect `json:"boundingBox"`

	// Contents of the LayoutText, if any. (optional)
	LayoutText string `json:"layoutText,omitempty"`

	// The post-layout inline text nodes, if any. (optional)
	InlineTextNodes []*css.InlineTextBox `json:"inlineTextNodes,omitempty"`

	// Index into the <code>computedStyles</code> array returned by <code>getSnapshot</code>. (optional)
	StyleIndex int `json:"styleIndex,omitempty"`
}

// A subset of the full ComputedStyle as defined by the request whitelist.

type ComputedStyle struct {
	// Name/value pairs of computed style properties.
	Properties []*NameValue `json:"properties"`
}

// A name/value pair.

type NameValue struct {
	// Attribute/property name.
	Name string `json:"name"`

	// Attribute/property value.
	Value string `json:"value"`
}

type GetSnapshotRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns a document snapshot, including the full DOM tree of the root node (including iframes, template contents, and imported documents) in a flattened array, as well as layout and white-listed computed style information for the nodes. Shadow DOM in the returned DOM tree is flattened.
func (d *Client) GetSnapshot() *GetSnapshotRequest {
	return &GetSnapshotRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whitelist of computed styles to return.
func (r *GetSnapshotRequest) ComputedStyleWhitelist(v []string) *GetSnapshotRequest {
	r.opts["computedStyleWhitelist"] = v
	return r
}

type GetSnapshotResult struct {
	// The nodes in the DOM tree. The DOMNode at index 0 corresponds to the root document.
	DomNodes []*DOMNode `json:"domNodes"`

	// The nodes in the layout tree.
	LayoutTreeNodes []*LayoutTreeNode `json:"layoutTreeNodes"`

	// Whitelisted ComputedStyle properties for each node in the layout tree.
	ComputedStyles []*ComputedStyle `json:"computedStyles"`
}

func (r *GetSnapshotRequest) Do() (*GetSnapshotResult, error) {
	var result GetSnapshotResult
	err := r.client.Call("DOMSnapshot.getSnapshot", r.opts, &result)
	return &result, err
}

func init() {
}
