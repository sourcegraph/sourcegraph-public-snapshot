package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CommitSearchResults []CommitSearchResult

func (c *CommitSearchResults) UnmarshalJSON(b []byte) error {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(b, &rawMessages); err != nil {
		return err
	}

	var results []CommitSearchResult
	for _, rawMessage := range rawMessages {
		var t struct {
			Typename string `json:"__typename"`
		}
		if err := json.Unmarshal(rawMessage, &t); err != nil {
			return err
		}

		switch t.Typename {
		case "CommitSearchResult", "":
		default:
			return errors.Errorf("expected result type %q, got %q", "CommitSearchResult", t.Typename)
		}

		var csr CommitSearchResult
		if err := json.Unmarshal(rawMessage, &csr); err != nil {
			return err
		}

		results = append(results, csr)
	}
	*c = results
	return nil
}

func (c CommitSearchResults) Value() (driver.Value, error) {
	if c == nil {
		c = CommitSearchResults{}
	}
	return json.Marshal(c)
}

func (c *CommitSearchResults) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("expected scanner value of type []byte, got %T", value)
	}
	return json.Unmarshal(b, c)
}

type CommitSearchResult struct {
	Refs           []Ref              `json:"refs"`
	SourceRefs     []Ref              `json:"sourceRefs"`
	MessagePreview *HighlightedString `json:"messagePreview"`
	DiffPreview    *HighlightedString `json:"diffPreview"`
	Commit         Commit             `json:"commit"`
}

type Ref struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Prefix      string `json:"prefix"`
}

type HighlightedString struct {
	Value      string `json:"value"`
	Highlights []struct {
		Line      int `json:"line"`
		Character int `json:"character"`
		Length    int `json:"length"`
	} `json:"highlights"`
}

type Commit struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Oid       string    `json:"oid"`
	Message   string    `json:"message"`
	Author    Signature `json:"author"`
	Committer Signature `json:"committer"`
}

type Signature struct {
	Person struct {
		DisplayName string `json:"displayName"`
	} `json:"person"`
	Date string `json:"date"`
}
