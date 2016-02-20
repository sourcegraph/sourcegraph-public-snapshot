// IAL implements inline attributes.

package mmark

import (
	"bytes"
	"sort"
	"strings"
)

// Do we return an anchor= when we see an #id or an id=
var anchorOrID = "anchor"

// One or more of these can be attached to block elements
type inlineAttr struct {
	id    string            // #id
	class map[string]bool   // 0 or more .class
	attr  map[string]string // key=value pairs
}

func newInlineAttr() *inlineAttr {
	return &inlineAttr{class: make(map[string]bool), attr: make(map[string]string)}
}

// Parsing and thus detecting an IAL. Return a valid *IAL or nil.
// IAL can have #id, .class or key=value element seperated by spaces, that may be escaped
func (p *parser) isInlineAttr(data []byte) int {
	esc := false
	quote := false
	ialB := 0
	ial := newInlineAttr()
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ':
			if quote {
				continue
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				ialB = i
				continue
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				} else {
					// this is illegal in an IAL, discard the posibility
					return 0
				}
			}
			ialB = i
		case '"':
			if esc {
				esc = !esc
				continue
			}
			quote = !quote
		case '\\':
			esc = !esc
		case '}':
			if esc {
				esc = !esc
				continue
			}
			// if this is mainmatter, frontmatter, or backmatter it isn't an IAL.
			s := string(data[1:i])
			switch s {
			case "frontmatter":
				fallthrough
			case "mainmatter":
				fallthrough
			case "backmatter":
				return 0
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				return i + 1
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				} else {
					// this is illegal in an IAL, discard the posibility
					return 0
				}
			}
			p.ial = p.ial.add(ial)
			return i + 1
		default:
			esc = false
		}
	}
	return 0
}

func parseKeyValue(chunk []byte) (string, string) {
	chunks := bytes.SplitN(chunk, []byte{'='}, 2)
	if len(chunks) != 2 {
		return "", ""
	}
	chunks[1] = bytes.Replace(chunks[1], []byte{'"'}, nil, -1)
	return string(chunks[0]), string(chunks[1])
}

// Add IAL to another, overwriting the #id, collapsing classes and attributes
func (i *inlineAttr) add(j *inlineAttr) *inlineAttr {
	if i == nil {
		return j
	}
	if j.id != "" {
		i.id = j.id
	}
	for k, c := range j.class {
		i.class[k] = c
	}
	for k, a := range j.attr {
		i.attr[k] = a
	}
	return i
}

// String renders an IAL and returns a string that can be included in the tag:
// class="class" anchor="id" key="value". The string s has a space as the first character.k
func (i *inlineAttr) String() (s string) {
	if i == nil {
		return ""
	}

	// some fluff needed to make this all sorted.
	if i.id != "" {
		s = " " + anchorOrID + "=\"" + i.id + "\""
	}

	keys := make([]string, 0, len(i.class))
	for k, _ := range i.class {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		s += " class=\"" + strings.Join(keys, " ") + "\""
	}

	keys = keys[:0]
	for k, _ := range i.attr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attr := make([]string, len(keys))
	for j, k := range keys {
		v := i.attr[k]
		attr[j] = k + "=\"" + v + "\""
	}
	if len(keys) > 0 {
		s += " " + strings.Join(attr, " ")
	}
	return s
}

// GetOrDefaultAttr sets the value under key if is is not set or
// use the value already in there. The boolean returns indicates
// if the value has been overwritten.
func (i *inlineAttr) GetOrDefaultAttr(key, def string) bool {
	v := i.attr[key]
	if v != "" {
		return false
	}
	if def == "" {
		return false
	}
	i.attr[key] = def
	return true
}

// GetOrDefaulClass sets the class class. The boolean returns indicates
// if the value has been overwritten.
func (i *inlineAttr) GetOrDefaultClass(class string) bool {
	_, ok := i.class[class]
	i.class[class] = true
	return ok
}

// GetOrDefaultID sets the id in i if it is not set. The boolean
// indicates if the id as set in i.
func (i *inlineAttr) GetOrDefaultId(id string) bool {
	if i.id != "" {
		return false
	}
	if id == "" {
		return false
	}
	i.id = id
	return true
}

// This returning a " "  is not particularly nice...

// Key returns the value of a specific key as a ' key="value"' string. If not found
// an string containing a space is returned.
func (i *inlineAttr) Key(key string) string {
	if v, ok := i.attr[key]; ok {
		return " " + key + "=\"" + v + "\""
	}
	return " "
}

// Value returns the value of a specific key as value.  If not found
// an empty string is returned. TODO(miek): should be " " or change Key() above.
func (i *inlineAttr) Value(key string) string {
	if v, ok := i.attr[key]; ok {
		return v
	}
	return ""
}

// DropAttr will drop the attribute under key from i.
// The returned boolean indicates if the key was found in i.
func (i *inlineAttr) DropAttr(key string) bool {
	_, ok := i.attr[key]
	delete(i.attr, key)
	return ok
}

// KeepAttr will drop all attributes, except the ones listed under keys.
func (i *inlineAttr) KeepAttr(keys []string) {
	newattr := make(map[string]string)
	for _, k := range keys {
		if v, ok := i.attr[k]; ok {
			newattr[k] = v
		}
	}
	i.attr = newattr
}

// KeepClass will drop all classes, except the ones listed under keys.
func (i *inlineAttr) KeepClass(keys []string) {
	newclass := make(map[string]bool)
	for _, k := range keys {
		if v, ok := i.class[k]; ok {
			newclass[k] = v
		}
	}
	i.class = newclass
}
