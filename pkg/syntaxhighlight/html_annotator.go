package syntaxhighlight

import (
	"strings"

	"github.com/sourcegraph/annotate"
)

// HTMLConfig is an object responsible for producing CSS classes for lexer tokens
type HTMLConfig interface {
	// Returns CSS class to use for a given token
	GetTokenClass(token Token) string
}

// HTMLConfig implementation that produces Pygments-like CSS classes
// See http://pygments.org/
type pygmentsHTMLConfig struct{}

// CSS class to use is typically token's type name
func (self pygmentsHTMLConfig) GetTokenClass(token Token) string {
	ttype := token.Type
	if STANDARD_TYPES[ttype.Name] != nil {
		return ttype.Name
	}
	tokens := make([]string, 0)
	for ttype != nil {
		tokens = append([]string{ttype.Name}, tokens...)
		ttype = ttype.parent
	}
	return strings.Join(tokens, "-")
}

// HTMLConfig implementation that keeps mapping between token type and CSS class to use.
type PaletteHTMLConfig struct {
	// Map (token type name => CSS class
	Palette map[string]string
}

// If current palette defines mapping for a given token type, returns matching CSS class,
// otherwise tries with the parent token type until there is any. Returns empty strings as fallback
func (self PaletteHTMLConfig) GetTokenClass(token Token) string {
	ttype := token.Type
	for ttype != nil {
		ret := self.Palette[ttype.Name]
		if ret != "" {
			return ret
		}
		ttype = ttype.parent
	}
	return ""
}

// Pygments-like HTML config
var PygmentsHTMLConfig HTMLConfig

// Google Prettify-like HTML config
var GooglePrettifyHTMLConfig HTMLConfig

// Default (AKA Google Prettify-like) HTML config
var DefaultHTMLConfig HTMLConfig

func init() {
	PygmentsHTMLConfig = &pygmentsHTMLConfig{}
	// see https://code.google.com/p/google-code-prettify/
	GooglePrettifyHTMLConfig = &PaletteHTMLConfig{Palette: map[string]string{
		String.String():         "str",
		Keyword.String():        "kwd",
		Comment.String():        "com",
		Keyword_Type.String():   "typ",
		Literal.String():        "lit",
		Punctuation.String():    "pun",
		Operator.String():       "pun",
		Name.String():           "pln",
		Name_Tag.String():       "tag",
		Name_Attribute.String(): "atn",
		Number.String():         "dec",
	}}
	DefaultHTMLConfig = GooglePrettifyHTMLConfig
}

// HTML annotator that transforms Token into Annotation object with left and right parts that look like <span>, </span>
type HTMLAnnotator struct {
	// HTML configuration to produce CSS classes
	Config HTMLConfig
}

// Initializes HTML annotator (does nothing)
func (self *HTMLAnnotator) Init() error {
	return nil
}

// Shuts down HTML annotator (does nothing)
func (self *HTMLAnnotator) Done() error {
	return nil
}

// Transforms token to annotation
func (self *HTMLAnnotator) Annotate(token Token) (*annotate.Annotation, error) {
	return &annotate.Annotation{
		Start:     token.Offset,
		End:       token.Offset + len(token.Text),
		Left:      []byte(`<span class="` + self.Config.GetTokenClass(token) + `">`),
		Right:     []byte(`</span>`),
		WantInner: 0,
	}, nil
}

// Instantiates new HTML annotator object using given configuration
func NewHTMLAnnotator(config HTMLConfig) *HTMLAnnotator {
	return &HTMLAnnotator{Config: config}
}
