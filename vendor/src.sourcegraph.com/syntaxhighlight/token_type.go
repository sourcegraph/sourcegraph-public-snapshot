package syntaxhighlight

// Token type object
type TokenType struct {
	// Token type name (using pygments notation)
	Name string `json:"name,omitempty"`
	// Parent's type if any. For example Keyword_Constant has Keyword as a parent
	parent *TokenType
}

// Returns parent type if any
func (self *TokenType) Parent() *TokenType {
	return self.parent
}

// Returns string representation of token type (pygments-like name)
func (self *TokenType) String() string {
	return self.Name
}

// Constructs new token type from pygments-like name
func NewTokenType(name string) *TokenType {
	ret := &TokenType{Name: name}
	STANDARD_TYPES[name] = ret
	return ret
}

// Constructs new token type from (pygments-like name) and parent type
func NewTokenTypeParent(name string, parent *TokenType) *TokenType {
	ret := &TokenType{Name: name, parent: parent}
	STANDARD_TYPES[name] = ret
	return ret
}

var (

	// Lists standard token types
	STANDARD_TYPES = map[string]*TokenType{}

	Whitespace = NewTokenType(`w`)
	Escape     = NewTokenType(`esc`)
	Error      = NewTokenType(`err`)
	Other      = NewTokenType(`x`)

	Keyword             = NewTokenType(`k`)
	Keyword_Constant    = NewTokenTypeParent(`kc`, Keyword)
	Keyword_Declaration = NewTokenTypeParent(`kd`, Keyword)
	Keyword_Namespace   = NewTokenTypeParent(`kn`, Keyword)
	Keyword_Pseudo      = NewTokenTypeParent(`kp`, Keyword)
	Keyword_Reserved    = NewTokenTypeParent(`kr`, Keyword)
	Keyword_Type        = NewTokenTypeParent(`kt`, Keyword)

	Name                   = NewTokenType(`n`)
	Name_Attribute         = NewTokenTypeParent(`na`, Name)
	Name_Builtin           = NewTokenTypeParent(`nb`, Name)
	Name_Builtin_Pseudo    = NewTokenTypeParent(`bp`, Name)
	Name_Class             = NewTokenTypeParent(`nc`, Name)
	Name_Constant          = NewTokenTypeParent(`no`, Name)
	Name_Decorator         = NewTokenTypeParent(`nd`, Name)
	Name_Entity            = NewTokenTypeParent(`ni`, Name)
	Name_Exception         = NewTokenTypeParent(`ne`, Name)
	Name_Function          = NewTokenTypeParent(`nf`, Name)
	Name_Property          = NewTokenTypeParent(`py`, Name)
	Name_Label             = NewTokenTypeParent(`nl`, Name)
	Name_Namespace         = NewTokenTypeParent(`nn`, Name)
	Name_Other             = NewTokenTypeParent(`nx`, Name)
	Name_Tag               = NewTokenTypeParent(`nt`, Name)
	Name_Variable          = NewTokenTypeParent(`nv`, Name)
	Name_Variable_Class    = NewTokenTypeParent(`vc`, Name_Variable)
	Name_Variable_Global   = NewTokenTypeParent(`vg`, Name_Variable)
	Name_Variable_Instance = NewTokenTypeParent(`vi`, Name_Variable)

	Literal      = NewTokenType(`l`)
	Literal_Date = NewTokenTypeParent(`ld`, Literal)

	String          = NewTokenType(`s`)
	String_Backtick = NewTokenTypeParent(`sb`, String)
	String_Char     = NewTokenTypeParent(`sc`, String)
	String_Doc      = NewTokenTypeParent(`sd`, String)
	String_Double   = NewTokenTypeParent(`s2`, String)
	String_Escape   = NewTokenTypeParent(`se`, String)
	String_Heredoc  = NewTokenTypeParent(`sh`, String)
	String_Interpol = NewTokenTypeParent(`si`, String)
	String_Other    = NewTokenTypeParent(`sx`, String)
	String_Regex    = NewTokenTypeParent(`sr`, String)
	String_Single   = NewTokenTypeParent(`s1`, String)
	String_Symbol   = NewTokenTypeParent(`ss`, String)

	Number              = NewTokenType(`m`)
	Number_Bin          = NewTokenTypeParent(`mb`, Number)
	Number_Float        = NewTokenTypeParent(`mf`, Number)
	Number_Hex          = NewTokenTypeParent(`mh`, Number)
	Number_Integer      = NewTokenTypeParent(`mi`, Number)
	Number_Integer_Long = NewTokenTypeParent(`il`, Number_Integer)
	Number_Oct          = NewTokenTypeParent(`mo`, Number)

	Operator      = NewTokenType(`o`)
	Operator_Word = NewTokenTypeParent(`ow`, Operator)

	Punctuation = NewTokenType(`p`)

	Comment           = NewTokenType(`c`)
	Comment_Multiline = NewTokenTypeParent(`cm`, Comment)
	Comment_Preproc   = NewTokenTypeParent(`cp`, Comment)
	Comment_Single    = NewTokenTypeParent(`c1`, Comment)
	Comment_Special   = NewTokenTypeParent(`cs`, Comment)

	Generic            = NewTokenType(`g`)
	Generic_Deleted    = NewTokenTypeParent(`gd`, Generic)
	Generic_Emph       = NewTokenTypeParent(`ge`, Generic)
	Generic_Error      = NewTokenTypeParent(`gr`, Generic)
	Generic_Heading    = NewTokenTypeParent(`gh`, Generic)
	Generic_Inserted   = NewTokenTypeParent(`gi`, Generic)
	Generic_Output     = NewTokenTypeParent(`go`, Generic)
	Generic_Prompt     = NewTokenTypeParent(`gp`, Generic)
	Generic_Strong     = NewTokenTypeParent(`gs`, Generic)
	Generic_Subheading = NewTokenTypeParent(`gu`, Generic)
	Generic_Traceback  = NewTokenTypeParent(`gt`, Generic)
)
