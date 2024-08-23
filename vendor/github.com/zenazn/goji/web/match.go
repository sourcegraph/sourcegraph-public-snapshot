package web

// The key used to store route Matches in the Goji environment. If this key is
// present in the environment and contains a value of type Match, routing will
// not be performed, and the Match's Handler will be used instead.
const MatchKey = "goji.web.Match"

// Match is the type of routing matches. It is inserted into C.Env under
// MatchKey when the Mux.Router middleware is invoked. If MatchKey is present at
// route dispatch time, the Handler of the corresponding Match will be called
// instead of performing routing as usual.
//
// By computing a Match and inserting it into the Goji environment as part of a
// middleware stack (see Mux.Router, for instance), it is possible to customize
// Goji's routing behavior or replace it entirely.
type Match struct {
	// Pattern is the Pattern that matched during routing. Will be nil if no
	// route matched (Handler will be set to the Mux's NotFound handler)
	Pattern Pattern
	// The Handler corresponding to the matched pattern.
	Handler Handler
}

// GetMatch returns the Match stored in the Goji environment, or an empty Match
// if none exists (valid Matches always have a Handler property).
func GetMatch(c C) Match {
	if c.Env == nil {
		return Match{}
	}
	mi, ok := c.Env[MatchKey]
	if !ok {
		return Match{}
	}
	if m, ok := mi.(Match); ok {
		return m
	}
	return Match{}
}

// RawPattern returns the PatternType that was originally passed to ParsePattern
// or any of the HTTP method functions (Get, Post, etc.).
func (m Match) RawPattern() PatternType {
	switch v := m.Pattern.(type) {
	case regexpPattern:
		return v.re
	case stringPattern:
		return v.raw
	default:
		return v
	}
}

// RawHandler returns the HandlerType that was originally passed to the HTTP
// method functions (Get, Post, etc.).
func (m Match) RawHandler() HandlerType {
	switch v := m.Handler.(type) {
	case netHTTPHandlerWrap:
		return v.Handler
	case handlerFuncWrap:
		return v.fn
	case netHTTPHandlerFuncWrap:
		return v.fn
	default:
		return v
	}
}
