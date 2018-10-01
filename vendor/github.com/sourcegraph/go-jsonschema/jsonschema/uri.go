package jsonschema

import (
	"net/url"
	"strconv"
	"strings"
)

// An ID identifies a JSON Schema.
//
// The identifier consists of the base URI and any JSON Pointer reference-tokens (see [RFC
// 6091](https://tools.ietf.org/html/rfc6901)).
type ID struct {
	// Base is the base URI that the reference tokens are resolved relative to.
	Base *url.URL

	// ReferenceTokens describe a value in the JSON document identified by the base URI.
	//
	// See [RFC 6091](https://tools.ietf.org/html/rfc6901).
	ReferenceTokens []ReferenceToken
}

// ResolveReference returns a copy of id with the provided reference tokens appended.
func (id ID) ResolveReference(ref []ReferenceToken) ID {
	tmp := make([]ReferenceToken, len(id.ReferenceTokens)+len(ref))
	copy(tmp, id.ReferenceTokens)
	copy(tmp[len(id.ReferenceTokens):], ref)
	return ID{
		Base:            id.Base,
		ReferenceTokens: tmp,
	}
}

// URI returns the URI for the ID, resolving the reference tokens relative to the base URI.
//
// If id.Base and id.ReferenceTokens are both nil, it returns nil.
func (id ID) URI() *url.URL {
	if id.ReferenceTokens == nil {
		return id.Base // can be nil
	}

	uri := &url.URL{Fragment: "/" + EncodeReferenceTokens(id.ReferenceTokens)}
	if id.Base != nil && id.Base.Fragment != "" {
		uri.Fragment = id.Base.Fragment + uri.Fragment
	}
	if id.Base != nil {
		uri = id.Base.ResolveReference(uri)
	}
	return uri
}

func (id ID) String() string {
	uri := id.URI()
	if uri == nil {
		return ""
	}
	return uri.String()
}

// A ReferenceToken describes a one-level traversal in a JSON document. See [RFC
// 6091](https://tools.ietf.org/html/rfc6901).
type ReferenceToken struct {
	Name    string // dereference object's named property
	Keyword bool   // if Name != "", whether the Name is a JSON Schema keyword (e.g., "properties", "items", etc.)
	Index   int    // dereference array's index
}

// EncodeReferenceTokens encodes the reference tokens to a string.
//
// TODO(sqs): Fully implement the encoding specified in
// https://tools.ietf.org/html/rfc6901#section-6.
func EncodeReferenceTokens(tokens []ReferenceToken) string {
	parts := make([]string, len(tokens))
	for i, token := range tokens {
		var part string
		if token.Name != "" {
			part = token.Name
		} else {
			part = strconv.Itoa(token.Index)
		}
		parts[i] = part
	}
	return strings.Join(parts, "/")
}
