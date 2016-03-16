package syntaxhighlight

// Lexer factory produces lexers
type LexerFactory func() Lexer

var (
	// map (extension (for example, .go) => lexer factory to produce lexers capable to tokenize files
	// with a given extension)
	extensionsMap = map[string]LexerFactory{}
	// map (MIME type (for example, text/x-gosrc) => lexer factory to produce lexers capable to tokenize files
	// with a given MIME type)
	mimeTypesMap = map[string]LexerFactory{}
)

// Instantiates new lexer for a given extension (for example .go) or returns nil if there is no registered lexer
// capable to handle given extension
func NewLexerByExtension(extension string) Lexer {
	factory := extensionsMap[extension]
	if factory == nil {
		return nil
	}
	return factory()
}

// Instantiates new lexer for a given MIME type (for example text/x-gosrc) or returns nil
// if there is no registered lexer capable to handle given MIME type
func NewLexerByMimeType(mimeType string) Lexer {
	factory := mimeTypesMap[mimeType]
	if factory == nil {
		return nil
	}
	return factory()
}

// Generates array of tokens by calling lexer til EOF.
// Please note that returned array does not contains gaps between tokens (whitespaces) as annotate does
func GetTokens(lexer Lexer, source []byte) []Token {
	lexer.Init(source)
	ret := make([]Token, 0, 20)
	t := lexer.NextToken()
	for t != nil {
		ret = append(ret, *t)
		t = lexer.NextToken()
	}
	return ret
}

// Registers new lexer factory, creates associations (extension => lexer, MIME type => lexer factory)
func register(fileExtensions []string, mimeTypes []string, factory LexerFactory) {
	for _, extension := range fileExtensions {
		extensionsMap[extension] = factory
	}
	for _, mimeType := range mimeTypes {
		mimeTypesMap[mimeType] = factory
	}
}
