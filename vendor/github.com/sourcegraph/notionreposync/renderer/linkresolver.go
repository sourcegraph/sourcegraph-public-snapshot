package renderer

import "errors"

var ErrDiscardLink = errors.New("discard link")

// LinkResolver can be implemented to modify Markdown links.
type LinkResolver interface {
	// ResolveLink accepts a link and returns it as-is or modified as desired,
	// for example to resolve an appropriate absolute link to the relevant
	// resource (e.g. another Notion document or a blob view).
	//
	// If ErrDiscardLink is returned, the link is converted into a plain text
	// element.
	ResolveLink(link string) (string, error)
}

// noopLinkResolver returns all links as-is and unmodified. It should be used
// as the default LinkResolver.
type noopLinkResolver struct{}

func (noopLinkResolver) ResolveLink(link string) (string, error) { return link, nil }

// DiscardLinkResolver discards all links, using ErrDiscardLink to indicate
// all links should be rendered as plain text.
type DiscardLinkResolver struct{}

func (DiscardLinkResolver) ResolveLink(link string) (string, error) { return "", ErrDiscardLink }
