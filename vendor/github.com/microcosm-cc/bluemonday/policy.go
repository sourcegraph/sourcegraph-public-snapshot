// Copyright (c) 2014, David Kitchen <david@buro9.com>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of the organisation (Microcosm) nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package bluemonday

import (
	"net/url"
	"regexp"
	"strings"
)

// Policy encapsulates the whitelist of HTML elements and attributes that will
// be applied to the sanitised HTML.
//
// You should use bluemonday.NewPolicy() to create a blank policy as the
// unexported fields contain maps that need to be initialized.
type Policy struct {

	// Declares whether the maps have been initialized, used as a cheap check to
	// ensure that those using Policy{} directly won't cause nil pointer
	// exceptions
	initialized bool

	// Allows the <!DOCTYPE > tag to exist in the sanitized document
	allowDocType bool

	// When true, add rel="nofollow" to HTML anchors
	requireNoFollow bool

	// When true, add rel="nofollow" to HTML anchors
	// Will add for href="http://foo"
	// Will skip for href="/foo" or href="foo"
	requireNoFollowFullyQualifiedLinks bool

	// When true add target="_blank" to fully qualified links
	// Will add for href="http://foo"
	// Will skip for href="/foo" or href="foo"
	addTargetBlankToFullyQualifiedLinks bool

	// When true, URLs must be parseable by "net/url" url.Parse()
	requireParseableURLs bool

	// When true, u, _ := url.Parse("url"); !u.IsAbs() is permitted
	allowRelativeURLs bool

	// map[htmlElementName]map[htmlAttributeName]attrPolicy
	elsAndAttrs map[string]map[string]attrPolicy

	// map[htmlAttributeName]attrPolicy
	globalAttrs map[string]attrPolicy

	// If urlPolicy is nil, all URLs with matching schema are allowed.
	// Otherwise, only the URLs with matching schema and urlPolicy(url)
	// returning true are allowed.
	allowURLSchemes map[string]urlPolicy

	setOfElementsWithoutAttrs  map[string]struct{}
	setOfElementsToSkipContent map[string]struct{}
}

type attrPolicy struct {

	// optional pattern to match, when not nil the regexp needs to match
	// otherwise the attribute is removed
	regexp *regexp.Regexp
}

type attrPolicyBuilder struct {
	p *Policy

	attrNames []string
	regexp    *regexp.Regexp
}

type urlPolicy func(url *url.URL) (allowUrl bool)

// init initializes the maps if this has not been done already
func (p *Policy) init() {
	if !p.initialized {
		p.elsAndAttrs = make(map[string]map[string]attrPolicy)
		p.globalAttrs = make(map[string]attrPolicy)
		p.allowURLSchemes = make(map[string]urlPolicy)
		p.setOfElementsWithoutAttrs = make(map[string]struct{})
		p.setOfElementsToSkipContent = make(map[string]struct{})
		p.initialized = true
	}
}

// NewPolicy returns a blank policy with nothing whitelisted or permitted. This
// is the recommended way to start building a policy and you should now use
// AllowAttrs() and/or AllowElements() to construct the whitelist of HTML
// elements and attributes.
func NewPolicy() *Policy {

	p := Policy{}

	p.addDefaultElementsWithoutAttrs()
	p.addDefaultSkipElementContent()

	return &p
}

// AllowAttrs takes a range of HTML attribute names and returns an
// attribute policy builder that allows you to specify the pattern and scope of
// the whitelisted attribute.
//
// The attribute policy is only added to the core policy when either Globally()
// or OnElements(...) are called.
func (p *Policy) AllowAttrs(attrNames ...string) *attrPolicyBuilder {

	p.init()

	abp := attrPolicyBuilder{p: p}

	for _, attrName := range attrNames {
		abp.attrNames = append(abp.attrNames, strings.ToLower(attrName))
	}

	return &abp
}

// Matching allows a regular expression to be applied to a nascent attribute
// policy, and returns the attribute policy. Calling this more than once will
// replace the existing regexp.
func (abp *attrPolicyBuilder) Matching(regex *regexp.Regexp) *attrPolicyBuilder {

	abp.regexp = regex

	return abp
}

// OnElements will bind an attribute policy to a given range of HTML elements
// and return the updated policy
func (abp *attrPolicyBuilder) OnElements(elements ...string) *Policy {

	for _, element := range elements {
		element = strings.ToLower(element)

		for _, attr := range abp.attrNames {

			if _, ok := abp.p.elsAndAttrs[element]; !ok {
				abp.p.elsAndAttrs[element] = make(map[string]attrPolicy)
			}

			ap := attrPolicy{}
			if abp.regexp != nil {
				ap.regexp = abp.regexp
			}

			abp.p.elsAndAttrs[element][attr] = ap
		}
	}

	return abp.p
}

// Globally will bind an attribute policy to all HTML elements and return the
// updated policy
func (abp *attrPolicyBuilder) Globally() *Policy {

	for _, attr := range abp.attrNames {
		if _, ok := abp.p.globalAttrs[attr]; !ok {
			abp.p.globalAttrs[attr] = attrPolicy{}
		}

		ap := attrPolicy{}
		if abp.regexp != nil {
			ap.regexp = abp.regexp
		}

		abp.p.globalAttrs[attr] = ap
	}

	return abp.p
}

// AllowElements will append HTML elements to the whitelist without applying an
// attribute policy to those elements (the elements are permitted
// sans-attributes)
func (p *Policy) AllowElements(names ...string) *Policy {
	p.init()

	for _, element := range names {
		element = strings.ToLower(element)

		if _, ok := p.elsAndAttrs[element]; !ok {
			p.elsAndAttrs[element] = make(map[string]attrPolicy)
		}
	}

	return p
}

// RequireNoFollowOnLinks will result in all <a> tags having a rel="nofollow"
// added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) RequireNoFollowOnLinks(require bool) *Policy {

	p.requireNoFollow = require
	p.requireParseableURLs = true

	return p
}

// RequireNoFollowOnFullyQualifiedLinks will result in all <a> tags that point
// to a non-local destination (i.e. starts with a protocol and has a host)
// having a rel="nofollow" added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) RequireNoFollowOnFullyQualifiedLinks(require bool) *Policy {

	p.requireNoFollowFullyQualifiedLinks = require
	p.requireParseableURLs = true

	return p
}

// AddTargetBlankToFullyQualifiedLinks will result in all <a> tags that point
// to a non-local destination (i.e. starts with a protocol and has a host)
// having a target="_blank" added to them if one does not already exist
//
// Note: This requires p.RequireParseableURLs(true) and will enable it.
func (p *Policy) AddTargetBlankToFullyQualifiedLinks(require bool) *Policy {

	p.addTargetBlankToFullyQualifiedLinks = require
	p.requireParseableURLs = true

	return p
}

// RequireParseableURLs will result in all URLs requiring that they be parseable
// by "net/url" url.Parse()
// This applies to:
// - a.href
// - area.href
// - blockquote.cite
// - img.src
// - link.href
// - script.src
func (p *Policy) RequireParseableURLs(require bool) *Policy {

	p.requireParseableURLs = require

	return p
}

// AllowRelativeURLs enables RequireParseableURLs and then permits URLs that
// are parseable, have no schema information and url.IsAbs() returns false
// This permits local URLs
func (p *Policy) AllowRelativeURLs(require bool) *Policy {

	p.RequireParseableURLs(true)
	p.allowRelativeURLs = require

	return p
}

// AllowURLSchemes will append URL schemes to the whitelist
// Example: p.AllowURLSchemes("mailto", "http", "https")
func (p *Policy) AllowURLSchemes(schemes ...string) *Policy {
	p.init()

	p.RequireParseableURLs(true)

	for _, scheme := range schemes {
		scheme = strings.ToLower(scheme)

		// Allow all URLs with matching scheme.
		p.allowURLSchemes[scheme] = nil
	}

	return p
}

// AllowURLSchemeWithCustomPolicy will append URL schemes with
// a custom URL policy to the whitelist.
// Only the URLs with matching schema and urlPolicy(url)
// returning true will be allowed.
func (p *Policy) AllowURLSchemeWithCustomPolicy(
	scheme string,
	urlPolicy func(url *url.URL) (allowUrl bool),
) *Policy {

	p.init()

	p.RequireParseableURLs(true)

	scheme = strings.ToLower(scheme)

	p.allowURLSchemes[scheme] = urlPolicy

	return p
}

// AllowDocType states whether the HTML sanitised by the sanitizer is allowed to
// contain the HTML DocType tag: <!DOCTYPE HTML> or one of it's variants.
//
// The HTML spec only permits one doctype per document, and as you know how you
// are using the output of this, you know best as to whether we should ignore it
// (default) or not.
//
// If you are sanitizing a HTML fragment the default (false) is fine.
func (p *Policy) AllowDocType(allow bool) *Policy {

	p.allowDocType = allow

	return p
}

// addDefaultElementsWithoutAttrs adds the HTML elements that we know are valid
// without any attributes to an internal map.
// i.e. we know that <table> is valid, but <bdo> isn't valid as the "dir" attr
// is mandatory
func (p *Policy) addDefaultElementsWithoutAttrs() {
	p.init()

	p.setOfElementsWithoutAttrs["abbr"] = struct{}{}
	p.setOfElementsWithoutAttrs["acronym"] = struct{}{}
	p.setOfElementsWithoutAttrs["article"] = struct{}{}
	p.setOfElementsWithoutAttrs["aside"] = struct{}{}
	p.setOfElementsWithoutAttrs["audio"] = struct{}{}
	p.setOfElementsWithoutAttrs["b"] = struct{}{}
	p.setOfElementsWithoutAttrs["bdi"] = struct{}{}
	p.setOfElementsWithoutAttrs["blockquote"] = struct{}{}
	p.setOfElementsWithoutAttrs["body"] = struct{}{}
	p.setOfElementsWithoutAttrs["br"] = struct{}{}
	p.setOfElementsWithoutAttrs["button"] = struct{}{}
	p.setOfElementsWithoutAttrs["canvas"] = struct{}{}
	p.setOfElementsWithoutAttrs["caption"] = struct{}{}
	p.setOfElementsWithoutAttrs["cite"] = struct{}{}
	p.setOfElementsWithoutAttrs["code"] = struct{}{}
	p.setOfElementsWithoutAttrs["col"] = struct{}{}
	p.setOfElementsWithoutAttrs["colgroup"] = struct{}{}
	p.setOfElementsWithoutAttrs["datalist"] = struct{}{}
	p.setOfElementsWithoutAttrs["dd"] = struct{}{}
	p.setOfElementsWithoutAttrs["del"] = struct{}{}
	p.setOfElementsWithoutAttrs["details"] = struct{}{}
	p.setOfElementsWithoutAttrs["dfn"] = struct{}{}
	p.setOfElementsWithoutAttrs["div"] = struct{}{}
	p.setOfElementsWithoutAttrs["dl"] = struct{}{}
	p.setOfElementsWithoutAttrs["dt"] = struct{}{}
	p.setOfElementsWithoutAttrs["em"] = struct{}{}
	p.setOfElementsWithoutAttrs["fieldset"] = struct{}{}
	p.setOfElementsWithoutAttrs["figcaption"] = struct{}{}
	p.setOfElementsWithoutAttrs["figure"] = struct{}{}
	p.setOfElementsWithoutAttrs["footer"] = struct{}{}
	p.setOfElementsWithoutAttrs["h1"] = struct{}{}
	p.setOfElementsWithoutAttrs["h2"] = struct{}{}
	p.setOfElementsWithoutAttrs["h3"] = struct{}{}
	p.setOfElementsWithoutAttrs["h4"] = struct{}{}
	p.setOfElementsWithoutAttrs["h5"] = struct{}{}
	p.setOfElementsWithoutAttrs["h6"] = struct{}{}
	p.setOfElementsWithoutAttrs["head"] = struct{}{}
	p.setOfElementsWithoutAttrs["header"] = struct{}{}
	p.setOfElementsWithoutAttrs["hgroup"] = struct{}{}
	p.setOfElementsWithoutAttrs["hr"] = struct{}{}
	p.setOfElementsWithoutAttrs["html"] = struct{}{}
	p.setOfElementsWithoutAttrs["i"] = struct{}{}
	p.setOfElementsWithoutAttrs["ins"] = struct{}{}
	p.setOfElementsWithoutAttrs["kbd"] = struct{}{}
	p.setOfElementsWithoutAttrs["li"] = struct{}{}
	p.setOfElementsWithoutAttrs["mark"] = struct{}{}
	p.setOfElementsWithoutAttrs["nav"] = struct{}{}
	p.setOfElementsWithoutAttrs["ol"] = struct{}{}
	p.setOfElementsWithoutAttrs["optgroup"] = struct{}{}
	p.setOfElementsWithoutAttrs["option"] = struct{}{}
	p.setOfElementsWithoutAttrs["p"] = struct{}{}
	p.setOfElementsWithoutAttrs["pre"] = struct{}{}
	p.setOfElementsWithoutAttrs["q"] = struct{}{}
	p.setOfElementsWithoutAttrs["rp"] = struct{}{}
	p.setOfElementsWithoutAttrs["rt"] = struct{}{}
	p.setOfElementsWithoutAttrs["ruby"] = struct{}{}
	p.setOfElementsWithoutAttrs["s"] = struct{}{}
	p.setOfElementsWithoutAttrs["samp"] = struct{}{}
	p.setOfElementsWithoutAttrs["section"] = struct{}{}
	p.setOfElementsWithoutAttrs["select"] = struct{}{}
	p.setOfElementsWithoutAttrs["small"] = struct{}{}
	p.setOfElementsWithoutAttrs["span"] = struct{}{}
	p.setOfElementsWithoutAttrs["strike"] = struct{}{}
	p.setOfElementsWithoutAttrs["strong"] = struct{}{}
	p.setOfElementsWithoutAttrs["style"] = struct{}{}
	p.setOfElementsWithoutAttrs["sub"] = struct{}{}
	p.setOfElementsWithoutAttrs["summary"] = struct{}{}
	p.setOfElementsWithoutAttrs["sup"] = struct{}{}
	p.setOfElementsWithoutAttrs["svg"] = struct{}{}
	p.setOfElementsWithoutAttrs["table"] = struct{}{}
	p.setOfElementsWithoutAttrs["tbody"] = struct{}{}
	p.setOfElementsWithoutAttrs["td"] = struct{}{}
	p.setOfElementsWithoutAttrs["textarea"] = struct{}{}
	p.setOfElementsWithoutAttrs["tfoot"] = struct{}{}
	p.setOfElementsWithoutAttrs["th"] = struct{}{}
	p.setOfElementsWithoutAttrs["thead"] = struct{}{}
	p.setOfElementsWithoutAttrs["time"] = struct{}{}
	p.setOfElementsWithoutAttrs["tr"] = struct{}{}
	p.setOfElementsWithoutAttrs["tt"] = struct{}{}
	p.setOfElementsWithoutAttrs["u"] = struct{}{}
	p.setOfElementsWithoutAttrs["ul"] = struct{}{}
	p.setOfElementsWithoutAttrs["var"] = struct{}{}
	p.setOfElementsWithoutAttrs["video"] = struct{}{}
	p.setOfElementsWithoutAttrs["wbr"] = struct{}{}

}

// addDefaultSkipElementContent adds the HTML elements that we should skip
// rendering the character content of, if the element itself is not allowed.
// This is all character data that the end user would not normally see.
// i.e. if we exclude a <script> tag then we shouldn't render the JavaScript or
// anything else until we encounter the closing </script> tag.
func (p *Policy) addDefaultSkipElementContent() {
	p.init()

	p.setOfElementsToSkipContent["frame"] = struct{}{}
	p.setOfElementsToSkipContent["frameset"] = struct{}{}
	p.setOfElementsToSkipContent["iframe"] = struct{}{}
	p.setOfElementsToSkipContent["noembed"] = struct{}{}
	p.setOfElementsToSkipContent["noframes"] = struct{}{}
	p.setOfElementsToSkipContent["noscript"] = struct{}{}
	p.setOfElementsToSkipContent["nostyle"] = struct{}{}
	p.setOfElementsToSkipContent["object"] = struct{}{}
	p.setOfElementsToSkipContent["script"] = struct{}{}
	p.setOfElementsToSkipContent["style"] = struct{}{}
	p.setOfElementsToSkipContent["title"] = struct{}{}
}
