package ann

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// Link is a type of annotation that refers to an arbitrary URL
	// (typically pointing to an external web page).
	Link = "link"
)

// LinkURL parses and returns a's link URL, if a's type is Link and if
// its Data contains a valid URL (encoded as a JSON string).
func (a *Ann) LinkURL() (*url.URL, error) {
	if a.Type != Link {
		return nil, &ErrType{Expected: Link, Actual: a.Type, Op: "LinkURL"}
	}
	var urlStr string
	if err := json.Unmarshal(a.Data, &urlStr); err != nil {
		return nil, err
	}
	return url.Parse(urlStr)
}

// SetLinkURL sets a's Type to Link and Data to the JSON
// representation of the URL string. If the URL is invalid, an error
// is returned.
func (a *Ann) SetLinkURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	b, err := json.Marshal(u.String())
	if err != nil {
		return err
	}
	a.Type = Link
	a.Data = b
	return nil
}

// ErrType indicates that an operation performed on an annotation
// expected the annotation to be a different type (e.g., calling
// LinkURL on a non-link annotation).
type ErrType struct {
	Expected, Actual string // Expected and actual types
	Op               string // The name of the operation or method that was called
}

func (e *ErrType) Error() string {
	return fmt.Sprintf("%s called on annotation type %q, expected type %q", e.Op, e.Actual, e.Expected)
}

func (a *Ann) sortKey() string {
	return strings.Join([]string{a.Repo, a.CommitID, a.UnitType, a.Unit, a.Type, a.File, strconv.Itoa(int(a.Start)), strconv.Itoa(int(a.End))}, ":")
}

// Sorting

type Anns []*Ann

func (vs Anns) Len() int           { return len(vs) }
func (vs Anns) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Anns) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }
