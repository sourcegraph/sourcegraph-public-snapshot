// This domain exposes CSS read/write operations. All CSS objects (stylesheets, rules, and styles) have an associated <code>id</code> used in subsequent operations on the related object. Each object type has a specific <code>id</code> structure, and those are not interchangeable between objects of different kinds. CSS objects can be loaded using the <code>get*ForNode()</code> calls (which accept a DOM node id). A client can also discover all the existing stylesheets with the <code>getAllStyleSheets()</code> method (or keeping track of the <code>styleSheetAdded</code>/<code>styleSheetRemoved</code> events) and subsequently load the required stylesheet contents using the <code>getStyleSheet[Text]()</code> methods. (experimental)
package css

import (
	"github.com/neelance/cdp-go/rpc"

	"github.com/neelance/cdp-go/protocol/dom"
)

// This domain exposes CSS read/write operations. All CSS objects (stylesheets, rules, and styles) have an associated <code>id</code> used in subsequent operations on the related object. Each object type has a specific <code>id</code> structure, and those are not interchangeable between objects of different kinds. CSS objects can be loaded using the <code>get*ForNode()</code> calls (which accept a DOM node id). A client can also discover all the existing stylesheets with the <code>getAllStyleSheets()</code> method (or keeping track of the <code>styleSheetAdded</code>/<code>styleSheetRemoved</code> events) and subsequently load the required stylesheet contents using the <code>getStyleSheet[Text]()</code> methods. (experimental)
type Client struct {
	*rpc.Client
}

type StyleSheetId string

// Stylesheet type: "injected" for stylesheets injected via extension, "user-agent" for user-agent stylesheets, "inspector" for stylesheets created by the inspector (i.e. those holding the "via inspector" rules), "regular" for regular stylesheets.

type StyleSheetOrigin string

// CSS rule collection for a single pseudo style.

type PseudoElementMatches struct {
	// Pseudo element type.
	PseudoType dom.PseudoType `json:"pseudoType"`

	// Matches of CSS rules applicable to the pseudo style.
	Matches []*RuleMatch `json:"matches"`
}

// Inherited CSS rule collection from ancestor node.

type InheritedStyleEntry struct {
	// The ancestor node's inline style, if any, in the style inheritance chain. (optional)
	InlineStyle *CSSStyle `json:"inlineStyle,omitempty"`

	// Matches of CSS rules matching the ancestor node in the style inheritance chain.
	MatchedCSSRules []*RuleMatch `json:"matchedCSSRules"`
}

// Match data for a CSS rule.

type RuleMatch struct {
	// CSS rule in the match.
	Rule *CSSRule `json:"rule"`

	// Matching selector indices in the rule's selectorList selectors (0-based).
	MatchingSelectors []int `json:"matchingSelectors"`
}

// Data for a simple selector (these are delimited by commas in a selector list).

type Value struct {
	// Value text.
	Text string `json:"text"`

	// Value range in the underlying resource (if available). (optional)
	Range *SourceRange `json:"range,omitempty"`
}

// Selector list data.

type SelectorList struct {
	// Selectors in the list.
	Selectors []*Value `json:"selectors"`

	// Rule selector text.
	Text string `json:"text"`
}

// CSS stylesheet metainformation.

type CSSStyleSheetHeader struct {
	// The stylesheet identifier.
	StyleSheetId StyleSheetId `json:"styleSheetId"`

	// Owner frame identifier.
	FrameId string `json:"frameId"`

	// Stylesheet resource URL.
	SourceURL string `json:"sourceURL"`

	// URL of source map associated with the stylesheet (if any). (optional)
	SourceMapURL string `json:"sourceMapURL,omitempty"`

	// Stylesheet origin.
	Origin StyleSheetOrigin `json:"origin"`

	// Stylesheet title.
	Title string `json:"title"`

	// The backend id for the owner node of the stylesheet. (optional)
	OwnerNode dom.BackendNodeId `json:"ownerNode,omitempty"`

	// Denotes whether the stylesheet is disabled.
	Disabled bool `json:"disabled"`

	// Whether the sourceURL field value comes from the sourceURL comment. (optional)
	HasSourceURL bool `json:"hasSourceURL,omitempty"`

	// Whether this stylesheet is created for STYLE tag by parser. This flag is not set for document.written STYLE tags.
	IsInline bool `json:"isInline"`

	// Line offset of the stylesheet within the resource (zero based).
	StartLine float64 `json:"startLine"`

	// Column offset of the stylesheet within the resource (zero based).
	StartColumn float64 `json:"startColumn"`

	// Size of the content (in characters).
	Length float64 `json:"length"`
}

// CSS rule representation.

type CSSRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified stylesheet rules) this rule came from. (optional)
	StyleSheetId StyleSheetId `json:"styleSheetId,omitempty"`

	// Rule selector data.
	SelectorList *SelectorList `json:"selectorList"`

	// Parent stylesheet's origin.
	Origin StyleSheetOrigin `json:"origin"`

	// Associated style declaration.
	Style *CSSStyle `json:"style"`

	// Media list array (for rules involving media queries). The array enumerates media queries starting with the innermost one, going outwards. (optional)
	Media []*CSSMedia `json:"media,omitempty"`
}

// CSS coverage information. (experimental)

type RuleUsage struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified stylesheet rules) this rule came from.
	StyleSheetId StyleSheetId `json:"styleSheetId"`

	// Offset of the start of the rule (including selector) from the beginning of the stylesheet.
	StartOffset float64 `json:"startOffset"`

	// Offset of the end of the rule body from the beginning of the stylesheet.
	EndOffset float64 `json:"endOffset"`

	// Indicates whether the rule was actually used by some element in the page.
	Used bool `json:"used"`
}

// Text range within a resource. All numbers are zero-based.

type SourceRange struct {
	// Start line of range.
	StartLine int `json:"startLine"`

	// Start column of range (inclusive).
	StartColumn int `json:"startColumn"`

	// End line of range
	EndLine int `json:"endLine"`

	// End column of range (exclusive).
	EndColumn int `json:"endColumn"`
}

type ShorthandEntry struct {
	// Shorthand name.
	Name string `json:"name"`

	// Shorthand value.
	Value string `json:"value"`

	// Whether the property has "!important" annotation (implies <code>false</code> if absent). (optional)
	Important bool `json:"important,omitempty"`
}

type CSSComputedStyleProperty struct {
	// Computed style property name.
	Name string `json:"name"`

	// Computed style property value.
	Value string `json:"value"`
}

// CSS style representation.

type CSSStyle struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified stylesheet rules) this rule came from. (optional)
	StyleSheetId StyleSheetId `json:"styleSheetId,omitempty"`

	// CSS properties in the style.
	CssProperties []*CSSProperty `json:"cssProperties"`

	// Computed values for all shorthands found in the style.
	ShorthandEntries []*ShorthandEntry `json:"shorthandEntries"`

	// Style declaration text (if available). (optional)
	CssText string `json:"cssText,omitempty"`

	// Style declaration range in the enclosing stylesheet (if available). (optional)
	Range *SourceRange `json:"range,omitempty"`
}

// CSS property declaration data.

type CSSProperty struct {
	// The property name.
	Name string `json:"name"`

	// The property value.
	Value string `json:"value"`

	// Whether the property has "!important" annotation (implies <code>false</code> if absent). (optional)
	Important bool `json:"important,omitempty"`

	// Whether the property is implicit (implies <code>false</code> if absent). (optional)
	Implicit bool `json:"implicit,omitempty"`

	// The full property text as specified in the style. (optional)
	Text string `json:"text,omitempty"`

	// Whether the property is understood by the browser (implies <code>true</code> if absent). (optional)
	ParsedOk bool `json:"parsedOk,omitempty"`

	// Whether the property is disabled by the user (present for source-based properties only). (optional)
	Disabled bool `json:"disabled,omitempty"`

	// The entire property range in the enclosing style declaration (if available). (optional)
	Range *SourceRange `json:"range,omitempty"`
}

// CSS media rule descriptor.

type CSSMedia struct {
	// Media query text.
	Text string `json:"text"`

	// Source of the media query: "mediaRule" if specified by a @media rule, "importRule" if specified by an @import rule, "linkedSheet" if specified by a "media" attribute in a linked stylesheet's LINK tag, "inlineSheet" if specified by a "media" attribute in an inline stylesheet's STYLE tag.
	Source string `json:"source"`

	// URL of the document containing the media query description. (optional)
	SourceURL string `json:"sourceURL,omitempty"`

	// The associated rule (@media or @import) header range in the enclosing stylesheet (if available). (optional)
	Range *SourceRange `json:"range,omitempty"`

	// Identifier of the stylesheet containing this object (if exists). (optional)
	StyleSheetId StyleSheetId `json:"styleSheetId,omitempty"`

	// Array of media queries. (optional, experimental)
	MediaList []*MediaQuery `json:"mediaList,omitempty"`
}

// Media query descriptor. (experimental)

type MediaQuery struct {
	// Array of media query expressions.
	Expressions []*MediaQueryExpression `json:"expressions"`

	// Whether the media query condition is satisfied.
	Active bool `json:"active"`
}

// Media query expression descriptor. (experimental)

type MediaQueryExpression struct {
	// Media query expression value.
	Value float64 `json:"value"`

	// Media query expression units.
	Unit string `json:"unit"`

	// Media query expression feature.
	Feature string `json:"feature"`

	// The associated range of the value text in the enclosing stylesheet (if available). (optional)
	ValueRange *SourceRange `json:"valueRange,omitempty"`

	// Computed length of media query expression (if applicable). (optional)
	ComputedLength float64 `json:"computedLength,omitempty"`
}

// Information about amount of glyphs that were rendered with given font. (experimental)

type PlatformFontUsage struct {
	// Font's family name reported by platform.
	FamilyName string `json:"familyName"`

	// Indicates if the font was downloaded or resolved locally.
	IsCustomFont bool `json:"isCustomFont"`

	// Amount of glyphs that were rendered with this font.
	GlyphCount float64 `json:"glyphCount"`
}

// CSS keyframes rule representation.

type CSSKeyframesRule struct {
	// Animation name.
	AnimationName *Value `json:"animationName"`

	// List of keyframes.
	Keyframes []*CSSKeyframeRule `json:"keyframes"`
}

// CSS keyframe rule representation.

type CSSKeyframeRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified stylesheet rules) this rule came from. (optional)
	StyleSheetId StyleSheetId `json:"styleSheetId,omitempty"`

	// Parent stylesheet's origin.
	Origin StyleSheetOrigin `json:"origin"`

	// Associated key text.
	KeyText *Value `json:"keyText"`

	// Associated style declaration.
	Style *CSSStyle `json:"style"`
}

// A descriptor of operation to mutate style declaration text.

type StyleDeclarationEdit struct {
	// The css style sheet identifier.
	StyleSheetId StyleSheetId `json:"styleSheetId"`

	// The range of the style text in the enclosing stylesheet.
	Range *SourceRange `json:"range"`

	// New style text.
	Text string `json:"text"`
}

// Details of post layout rendered text positions. The exact layout should not be regarded as stable and may change between versions. (experimental)

type InlineTextBox struct {
	// The absolute position bounding box.
	BoundingBox *dom.Rect `json:"boundingBox"`

	// The starting index in characters, for this post layout textbox substring.
	StartCharacterIndex int `json:"startCharacterIndex"`

	// The number of characters in this post layout textbox substring.
	NumCharacters int `json:"numCharacters"`
}

// Details of an element in the DOM tree with a LayoutObject. (experimental)

type LayoutTreeNode struct {
	// The id of the related DOM node matching one from DOM.GetDocument.
	NodeId dom.NodeId `json:"nodeId"`

	// The absolute position bounding box.
	BoundingBox *dom.Rect `json:"boundingBox"`

	// Contents of the LayoutText if any (optional)
	LayoutText string `json:"layoutText,omitempty"`

	// The post layout inline text nodes, if any. (optional)
	InlineTextNodes []*InlineTextBox `json:"inlineTextNodes,omitempty"`

	// Index into the computedStyles array returned by getLayoutTreeAndStyles. (optional)
	StyleIndex int `json:"styleIndex,omitempty"`
}

// A subset of the full ComputedStyle as defined by the request whitelist. (experimental)

type ComputedStyle struct {
	Properties []*CSSComputedStyleProperty `json:"properties"`
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables the CSS agent for the given page. Clients should not assume that the CSS agent has been enabled until the result of this command is received.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("CSS.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables the CSS agent for the given page.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("CSS.disable", r.opts, nil)
}

type GetMatchedStylesForNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns requested styles for a DOM node identified by <code>nodeId</code>.
func (d *Client) GetMatchedStylesForNode() *GetMatchedStylesForNodeRequest {
	return &GetMatchedStylesForNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetMatchedStylesForNodeRequest) NodeId(v dom.NodeId) *GetMatchedStylesForNodeRequest {
	r.opts["nodeId"] = v
	return r
}

type GetMatchedStylesForNodeResult struct {
	// Inline style for the specified DOM node. (optional)
	InlineStyle *CSSStyle `json:"inlineStyle"`

	// Attribute-defined element style (e.g. resulting from "width=20 height=100%"). (optional)
	AttributesStyle *CSSStyle `json:"attributesStyle"`

	// CSS rules matching this node, from all applicable stylesheets. (optional)
	MatchedCSSRules []*RuleMatch `json:"matchedCSSRules"`

	// Pseudo style matches for this node. (optional)
	PseudoElements []*PseudoElementMatches `json:"pseudoElements"`

	// A chain of inherited styles (from the immediate node parent up to the DOM tree root). (optional)
	Inherited []*InheritedStyleEntry `json:"inherited"`

	// A list of CSS keyframed animations matching this node. (optional)
	CssKeyframesRules []*CSSKeyframesRule `json:"cssKeyframesRules"`
}

func (r *GetMatchedStylesForNodeRequest) Do() (*GetMatchedStylesForNodeResult, error) {
	var result GetMatchedStylesForNodeResult
	err := r.client.Call("CSS.getMatchedStylesForNode", r.opts, &result)
	return &result, err
}

type GetInlineStylesForNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the styles defined inline (explicitly in the "style" attribute and implicitly, using DOM attributes) for a DOM node identified by <code>nodeId</code>.
func (d *Client) GetInlineStylesForNode() *GetInlineStylesForNodeRequest {
	return &GetInlineStylesForNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetInlineStylesForNodeRequest) NodeId(v dom.NodeId) *GetInlineStylesForNodeRequest {
	r.opts["nodeId"] = v
	return r
}

type GetInlineStylesForNodeResult struct {
	// Inline style for the specified DOM node. (optional)
	InlineStyle *CSSStyle `json:"inlineStyle"`

	// Attribute-defined element style (e.g. resulting from "width=20 height=100%"). (optional)
	AttributesStyle *CSSStyle `json:"attributesStyle"`
}

func (r *GetInlineStylesForNodeRequest) Do() (*GetInlineStylesForNodeResult, error) {
	var result GetInlineStylesForNodeResult
	err := r.client.Call("CSS.getInlineStylesForNode", r.opts, &result)
	return &result, err
}

type GetComputedStyleForNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the computed style for a DOM node identified by <code>nodeId</code>.
func (d *Client) GetComputedStyleForNode() *GetComputedStyleForNodeRequest {
	return &GetComputedStyleForNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetComputedStyleForNodeRequest) NodeId(v dom.NodeId) *GetComputedStyleForNodeRequest {
	r.opts["nodeId"] = v
	return r
}

type GetComputedStyleForNodeResult struct {
	// Computed style for the specified DOM node.
	ComputedStyle []*CSSComputedStyleProperty `json:"computedStyle"`
}

func (r *GetComputedStyleForNodeRequest) Do() (*GetComputedStyleForNodeResult, error) {
	var result GetComputedStyleForNodeResult
	err := r.client.Call("CSS.getComputedStyleForNode", r.opts, &result)
	return &result, err
}

type GetPlatformFontsForNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Requests information about platform fonts which we used to render child TextNodes in the given node. (experimental)
func (d *Client) GetPlatformFontsForNode() *GetPlatformFontsForNodeRequest {
	return &GetPlatformFontsForNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetPlatformFontsForNodeRequest) NodeId(v dom.NodeId) *GetPlatformFontsForNodeRequest {
	r.opts["nodeId"] = v
	return r
}

type GetPlatformFontsForNodeResult struct {
	// Usage statistics for every employed platform font.
	Fonts []*PlatformFontUsage `json:"fonts"`
}

func (r *GetPlatformFontsForNodeRequest) Do() (*GetPlatformFontsForNodeResult, error) {
	var result GetPlatformFontsForNodeResult
	err := r.client.Call("CSS.getPlatformFontsForNode", r.opts, &result)
	return &result, err
}

type GetStyleSheetTextRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns the current textual content and the URL for a stylesheet.
func (d *Client) GetStyleSheetText() *GetStyleSheetTextRequest {
	return &GetStyleSheetTextRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *GetStyleSheetTextRequest) StyleSheetId(v StyleSheetId) *GetStyleSheetTextRequest {
	r.opts["styleSheetId"] = v
	return r
}

type GetStyleSheetTextResult struct {
	// The stylesheet text.
	Text string `json:"text"`
}

func (r *GetStyleSheetTextRequest) Do() (*GetStyleSheetTextResult, error) {
	var result GetStyleSheetTextResult
	err := r.client.Call("CSS.getStyleSheetText", r.opts, &result)
	return &result, err
}

type CollectClassNamesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns all class names from specified stylesheet. (experimental)
func (d *Client) CollectClassNames() *CollectClassNamesRequest {
	return &CollectClassNamesRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *CollectClassNamesRequest) StyleSheetId(v StyleSheetId) *CollectClassNamesRequest {
	r.opts["styleSheetId"] = v
	return r
}

type CollectClassNamesResult struct {
	// Class name list.
	ClassNames []string `json:"classNames"`
}

func (r *CollectClassNamesRequest) Do() (*CollectClassNamesResult, error) {
	var result CollectClassNamesResult
	err := r.client.Call("CSS.collectClassNames", r.opts, &result)
	return &result, err
}

type SetStyleSheetTextRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Sets the new stylesheet text.
func (d *Client) SetStyleSheetText() *SetStyleSheetTextRequest {
	return &SetStyleSheetTextRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetStyleSheetTextRequest) StyleSheetId(v StyleSheetId) *SetStyleSheetTextRequest {
	r.opts["styleSheetId"] = v
	return r
}

func (r *SetStyleSheetTextRequest) Text(v string) *SetStyleSheetTextRequest {
	r.opts["text"] = v
	return r
}

type SetStyleSheetTextResult struct {
	// URL of source map associated with script (if any). (optional)
	SourceMapURL string `json:"sourceMapURL"`
}

func (r *SetStyleSheetTextRequest) Do() (*SetStyleSheetTextResult, error) {
	var result SetStyleSheetTextResult
	err := r.client.Call("CSS.setStyleSheetText", r.opts, &result)
	return &result, err
}

type SetRuleSelectorRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Modifies the rule selector.
func (d *Client) SetRuleSelector() *SetRuleSelectorRequest {
	return &SetRuleSelectorRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetRuleSelectorRequest) StyleSheetId(v StyleSheetId) *SetRuleSelectorRequest {
	r.opts["styleSheetId"] = v
	return r
}

func (r *SetRuleSelectorRequest) Range(v *SourceRange) *SetRuleSelectorRequest {
	r.opts["range"] = v
	return r
}

func (r *SetRuleSelectorRequest) Selector(v string) *SetRuleSelectorRequest {
	r.opts["selector"] = v
	return r
}

type SetRuleSelectorResult struct {
	// The resulting selector list after modification.
	SelectorList *SelectorList `json:"selectorList"`
}

func (r *SetRuleSelectorRequest) Do() (*SetRuleSelectorResult, error) {
	var result SetRuleSelectorResult
	err := r.client.Call("CSS.setRuleSelector", r.opts, &result)
	return &result, err
}

type SetKeyframeKeyRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Modifies the keyframe rule key text.
func (d *Client) SetKeyframeKey() *SetKeyframeKeyRequest {
	return &SetKeyframeKeyRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetKeyframeKeyRequest) StyleSheetId(v StyleSheetId) *SetKeyframeKeyRequest {
	r.opts["styleSheetId"] = v
	return r
}

func (r *SetKeyframeKeyRequest) Range(v *SourceRange) *SetKeyframeKeyRequest {
	r.opts["range"] = v
	return r
}

func (r *SetKeyframeKeyRequest) KeyText(v string) *SetKeyframeKeyRequest {
	r.opts["keyText"] = v
	return r
}

type SetKeyframeKeyResult struct {
	// The resulting key text after modification.
	KeyText *Value `json:"keyText"`
}

func (r *SetKeyframeKeyRequest) Do() (*SetKeyframeKeyResult, error) {
	var result SetKeyframeKeyResult
	err := r.client.Call("CSS.setKeyframeKey", r.opts, &result)
	return &result, err
}

type SetStyleTextsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Applies specified style edits one after another in the given order.
func (d *Client) SetStyleTexts() *SetStyleTextsRequest {
	return &SetStyleTextsRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetStyleTextsRequest) Edits(v []*StyleDeclarationEdit) *SetStyleTextsRequest {
	r.opts["edits"] = v
	return r
}

type SetStyleTextsResult struct {
	// The resulting styles after modification.
	Styles []*CSSStyle `json:"styles"`
}

func (r *SetStyleTextsRequest) Do() (*SetStyleTextsResult, error) {
	var result SetStyleTextsResult
	err := r.client.Call("CSS.setStyleTexts", r.opts, &result)
	return &result, err
}

type SetMediaTextRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Modifies the rule selector.
func (d *Client) SetMediaText() *SetMediaTextRequest {
	return &SetMediaTextRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetMediaTextRequest) StyleSheetId(v StyleSheetId) *SetMediaTextRequest {
	r.opts["styleSheetId"] = v
	return r
}

func (r *SetMediaTextRequest) Range(v *SourceRange) *SetMediaTextRequest {
	r.opts["range"] = v
	return r
}

func (r *SetMediaTextRequest) Text(v string) *SetMediaTextRequest {
	r.opts["text"] = v
	return r
}

type SetMediaTextResult struct {
	// The resulting CSS media rule after modification.
	Media *CSSMedia `json:"media"`
}

func (r *SetMediaTextRequest) Do() (*SetMediaTextResult, error) {
	var result SetMediaTextResult
	err := r.client.Call("CSS.setMediaText", r.opts, &result)
	return &result, err
}

type CreateStyleSheetRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Creates a new special "via-inspector" stylesheet in the frame with given <code>frameId</code>.
func (d *Client) CreateStyleSheet() *CreateStyleSheetRequest {
	return &CreateStyleSheetRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the frame where "via-inspector" stylesheet should be created.
func (r *CreateStyleSheetRequest) FrameId(v string) *CreateStyleSheetRequest {
	r.opts["frameId"] = v
	return r
}

type CreateStyleSheetResult struct {
	// Identifier of the created "via-inspector" stylesheet.
	StyleSheetId StyleSheetId `json:"styleSheetId"`
}

func (r *CreateStyleSheetRequest) Do() (*CreateStyleSheetResult, error) {
	var result CreateStyleSheetResult
	err := r.client.Call("CSS.createStyleSheet", r.opts, &result)
	return &result, err
}

type AddRuleRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Inserts a new rule with the given <code>ruleText</code> in a stylesheet with given <code>styleSheetId</code>, at the position specified by <code>location</code>.
func (d *Client) AddRule() *AddRuleRequest {
	return &AddRuleRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The css style sheet identifier where a new rule should be inserted.
func (r *AddRuleRequest) StyleSheetId(v StyleSheetId) *AddRuleRequest {
	r.opts["styleSheetId"] = v
	return r
}

// The text of a new rule.
func (r *AddRuleRequest) RuleText(v string) *AddRuleRequest {
	r.opts["ruleText"] = v
	return r
}

// Text position of a new rule in the target style sheet.
func (r *AddRuleRequest) Location(v *SourceRange) *AddRuleRequest {
	r.opts["location"] = v
	return r
}

type AddRuleResult struct {
	// The newly created rule.
	Rule *CSSRule `json:"rule"`
}

func (r *AddRuleRequest) Do() (*AddRuleResult, error) {
	var result AddRuleResult
	err := r.client.Call("CSS.addRule", r.opts, &result)
	return &result, err
}

type ForcePseudoStateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Ensures that the given node will have specified pseudo-classes whenever its style is computed by the browser.
func (d *Client) ForcePseudoState() *ForcePseudoStateRequest {
	return &ForcePseudoStateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The element id for which to force the pseudo state.
func (r *ForcePseudoStateRequest) NodeId(v dom.NodeId) *ForcePseudoStateRequest {
	r.opts["nodeId"] = v
	return r
}

// Element pseudo classes to force when computing the element's style.
func (r *ForcePseudoStateRequest) ForcedPseudoClasses(v []string) *ForcePseudoStateRequest {
	r.opts["forcedPseudoClasses"] = v
	return r
}

func (r *ForcePseudoStateRequest) Do() error {
	return r.client.Call("CSS.forcePseudoState", r.opts, nil)
}

type GetMediaQueriesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns all media queries parsed by the rendering engine. (experimental)
func (d *Client) GetMediaQueries() *GetMediaQueriesRequest {
	return &GetMediaQueriesRequest{opts: make(map[string]interface{}), client: d.Client}
}

type GetMediaQueriesResult struct {
	Medias []*CSSMedia `json:"medias"`
}

func (r *GetMediaQueriesRequest) Do() (*GetMediaQueriesResult, error) {
	var result GetMediaQueriesResult
	err := r.client.Call("CSS.getMediaQueries", r.opts, &result)
	return &result, err
}

type SetEffectivePropertyValueForNodeRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Find a rule with the given active property for the given node and set the new value for this property (experimental)
func (d *Client) SetEffectivePropertyValueForNode() *SetEffectivePropertyValueForNodeRequest {
	return &SetEffectivePropertyValueForNodeRequest{opts: make(map[string]interface{}), client: d.Client}
}

// The element id for which to set property.
func (r *SetEffectivePropertyValueForNodeRequest) NodeId(v dom.NodeId) *SetEffectivePropertyValueForNodeRequest {
	r.opts["nodeId"] = v
	return r
}

func (r *SetEffectivePropertyValueForNodeRequest) PropertyName(v string) *SetEffectivePropertyValueForNodeRequest {
	r.opts["propertyName"] = v
	return r
}

func (r *SetEffectivePropertyValueForNodeRequest) Value(v string) *SetEffectivePropertyValueForNodeRequest {
	r.opts["value"] = v
	return r
}

func (r *SetEffectivePropertyValueForNodeRequest) Do() error {
	return r.client.Call("CSS.setEffectivePropertyValueForNode", r.opts, nil)
}

type GetBackgroundColorsRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) GetBackgroundColors() *GetBackgroundColorsRequest {
	return &GetBackgroundColorsRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of the node to get background colors for.
func (r *GetBackgroundColorsRequest) NodeId(v dom.NodeId) *GetBackgroundColorsRequest {
	r.opts["nodeId"] = v
	return r
}

type GetBackgroundColorsResult struct {
	// The range of background colors behind this element, if it contains any visible text. If no visible text is present, this will be undefined. In the case of a flat background color, this will consist of simply that color. In the case of a gradient, this will consist of each of the color stops. For anything more complicated, this will be an empty array. Images will be ignored (as if the image had failed to load). (optional)
	BackgroundColors []string `json:"backgroundColors"`
}

func (r *GetBackgroundColorsRequest) Do() (*GetBackgroundColorsResult, error) {
	var result GetBackgroundColorsResult
	err := r.client.Call("CSS.getBackgroundColors", r.opts, &result)
	return &result, err
}

type GetLayoutTreeAndStylesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// For the main document and any content documents, return the LayoutTreeNodes and a whitelisted subset of the computed style. It only returns pushed nodes, on way to pull all nodes is to call DOM.getDocument with a depth of -1. (experimental)
func (d *Client) GetLayoutTreeAndStyles() *GetLayoutTreeAndStylesRequest {
	return &GetLayoutTreeAndStylesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Whitelist of computed styles to return.
func (r *GetLayoutTreeAndStylesRequest) ComputedStyleWhitelist(v []string) *GetLayoutTreeAndStylesRequest {
	r.opts["computedStyleWhitelist"] = v
	return r
}

type GetLayoutTreeAndStylesResult struct {
	LayoutTreeNodes []*LayoutTreeNode `json:"layoutTreeNodes"`

	ComputedStyles []*ComputedStyle `json:"computedStyles"`
}

func (r *GetLayoutTreeAndStylesRequest) Do() (*GetLayoutTreeAndStylesResult, error) {
	var result GetLayoutTreeAndStylesResult
	err := r.client.Call("CSS.getLayoutTreeAndStyles", r.opts, &result)
	return &result, err
}

type StartRuleUsageTrackingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables the selector recording. (experimental)
func (d *Client) StartRuleUsageTracking() *StartRuleUsageTrackingRequest {
	return &StartRuleUsageTrackingRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *StartRuleUsageTrackingRequest) Do() error {
	return r.client.Call("CSS.startRuleUsageTracking", r.opts, nil)
}

type TakeCoverageDeltaRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Obtain list of rules that became used since last call to this method (or since start of coverage instrumentation) (experimental)
func (d *Client) TakeCoverageDelta() *TakeCoverageDeltaRequest {
	return &TakeCoverageDeltaRequest{opts: make(map[string]interface{}), client: d.Client}
}

type TakeCoverageDeltaResult struct {
	Coverage []*RuleUsage `json:"coverage"`
}

func (r *TakeCoverageDeltaRequest) Do() (*TakeCoverageDeltaResult, error) {
	var result TakeCoverageDeltaResult
	err := r.client.Call("CSS.takeCoverageDelta", r.opts, &result)
	return &result, err
}

type StopRuleUsageTrackingRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// The list of rules with an indication of whether these were used (experimental)
func (d *Client) StopRuleUsageTracking() *StopRuleUsageTrackingRequest {
	return &StopRuleUsageTrackingRequest{opts: make(map[string]interface{}), client: d.Client}
}

type StopRuleUsageTrackingResult struct {
	RuleUsage []*RuleUsage `json:"ruleUsage"`
}

func (r *StopRuleUsageTrackingRequest) Do() (*StopRuleUsageTrackingResult, error) {
	var result StopRuleUsageTrackingResult
	err := r.client.Call("CSS.stopRuleUsageTracking", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["CSS.mediaQueryResultChanged"] = func() interface{} { return new(MediaQueryResultChangedEvent) }
	rpc.EventTypes["CSS.fontsUpdated"] = func() interface{} { return new(FontsUpdatedEvent) }
	rpc.EventTypes["CSS.styleSheetChanged"] = func() interface{} { return new(StyleSheetChangedEvent) }
	rpc.EventTypes["CSS.styleSheetAdded"] = func() interface{} { return new(StyleSheetAddedEvent) }
	rpc.EventTypes["CSS.styleSheetRemoved"] = func() interface{} { return new(StyleSheetRemovedEvent) }
}

// Fires whenever a MediaQuery result changes (for example, after a browser window has been resized.) The current implementation considers only viewport-dependent media features.
type MediaQueryResultChangedEvent struct {
}

// Fires whenever a web font gets loaded.
type FontsUpdatedEvent struct {
}

// Fired whenever a stylesheet is changed as a result of the client operation.
type StyleSheetChangedEvent struct {
	StyleSheetId StyleSheetId `json:"styleSheetId"`
}

// Fired whenever an active document stylesheet is added.
type StyleSheetAddedEvent struct {
	// Added stylesheet metainfo.
	Header *CSSStyleSheetHeader `json:"header"`
}

// Fired whenever an active document stylesheet is removed.
type StyleSheetRemovedEvent struct {
	// Identifier of the removed stylesheet.
	StyleSheetId StyleSheetId `json:"styleSheetId"`
}
