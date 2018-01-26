// (experimental)
package accessibility

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
)

// (experimental)
type Client struct {
	*rpc.Client
}

// Unique accessibility node identifier.

type AXNodeId string

// Enum of possible property types.

type AXValueType string

// Enum of possible property sources.

type AXValueSourceType string

// Enum of possible native property sources (as a subtype of a particular AXValueSourceType).

type AXValueNativeSourceType string

// A single source for a computed AX property.

type AXValueSource struct {
	// What type of source this is.
	Type AXValueSourceType `json:"type"`

	// The value of this property source. (optional)
	Value *AXValue `json:"value,omitempty"`

	// The name of the relevant attribute, if any. (optional)
	Attribute string `json:"attribute,omitempty"`

	// The value of the relevant attribute, if any. (optional)
	AttributeValue *AXValue `json:"attributeValue,omitempty"`

	// Whether this source is superseded by a higher priority source. (optional)
	Superseded bool `json:"superseded,omitempty"`

	// The native markup source for this value, e.g. a <label> element. (optional)
	NativeSource AXValueNativeSourceType `json:"nativeSource,omitempty"`

	// The value, such as a node or node list, of the native source. (optional)
	NativeSourceValue *AXValue `json:"nativeSourceValue,omitempty"`

	// Whether the value for this property is invalid. (optional)
	Invalid bool `json:"invalid,omitempty"`

	// Reason for the value being invalid, if it is. (optional)
	InvalidReason string `json:"invalidReason,omitempty"`
}

type AXRelatedNode struct {
	// The BackendNodeId of the related DOM node.
	BackendDOMNodeId dom.BackendNodeId `json:"backendDOMNodeId"`

	// The IDRef value provided, if any. (optional)
	Idref string `json:"idref,omitempty"`

	// The text alternative of this node in the current context. (optional)
	Text string `json:"text,omitempty"`
}

type AXProperty struct {
	// The name of this property.
	Name string `json:"name"`

	// The value of this property.
	Value *AXValue `json:"value"`
}

// A single computed AX property.

type AXValue struct {
	// The type of this value.
	Type AXValueType `json:"type"`

	// The computed value of this property. (optional)
	Value interface{} `json:"value,omitempty"`

	// One or more related nodes, if applicable. (optional)
	RelatedNodes []*AXRelatedNode `json:"relatedNodes,omitempty"`

	// The sources which contributed to the computation of this property. (optional)
	Sources []*AXValueSource `json:"sources,omitempty"`
}

// States which apply to every AX node.

type AXGlobalStates string

// Attributes which apply to nodes in live regions.

type AXLiveRegionAttributes string

// Attributes which apply to widgets.

type AXWidgetAttributes string

// States which apply to widgets.

type AXWidgetStates string

// Relationships between elements other than parent/child/sibling.

type AXRelationshipAttributes string

// A node in the accessibility tree.

type AXNode struct {
	// Unique identifier for this node.
	NodeId AXNodeId `json:"nodeId"`

	// Whether this node is ignored for accessibility
	Ignored bool `json:"ignored"`

	// Collection of reasons why this node is hidden. (optional)
	IgnoredReasons []*AXProperty `json:"ignoredReasons,omitempty"`

	// This <code>Node</code>'s role, whether explicit or implicit. (optional)
	Role *AXValue `json:"role,omitempty"`

	// The accessible name for this <code>Node</code>. (optional)
	Name *AXValue `json:"name,omitempty"`

	// The accessible description for this <code>Node</code>. (optional)
	Description *AXValue `json:"description,omitempty"`

	// The value for this <code>Node</code>. (optional)
	Value *AXValue `json:"value,omitempty"`

	// All other properties (optional)
	Properties []*AXProperty `json:"properties,omitempty"`

	// IDs for each of this node's child nodes. (optional)
	ChildIds []AXNodeId `json:"childIds,omitempty"`

	// The backend ID for the associated DOM node, if any. (optional)
	BackendDOMNodeId dom.BackendNodeId `json:"backendDOMNodeId,omitempty"`
}

type GetPartialAXTreeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Fetches the accessibility node and partial accessibility tree for this DOM node, if it exists. (experimental)
func (d *Client) GetPartialAXTree() *GetPartialAXTreeRequest {
	return &GetPartialAXTreeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// ID of node to get the partial accessibility tree for.
func (r *GetPartialAXTreeRequest) NodeId(v dom.NodeId) *GetPartialAXTreeRequest {
	r.opts["nodeId"] = v
	return r
}

// Whether to fetch this nodes ancestors, siblings and children. Defaults to true. (optional)
func (r *GetPartialAXTreeRequest) FetchRelatives(v bool) *GetPartialAXTreeRequest {
	r.opts["fetchRelatives"] = v
	return r
}

type GetPartialAXTreeResult struct {
	// The <code>Accessibility.AXNode</code> for this DOM node, if it exists, plus its ancestors, siblings and children, if requested.
	Nodes []*AXNode `json:"nodes"`
}

func (r *GetPartialAXTreeRequest) Do() (*GetPartialAXTreeResult, error) {
	var result GetPartialAXTreeResult
	err := r.client.Call("Accessibility.getPartialAXTree", r.opts, &result)
	return &result, err
}

func init() {
}
