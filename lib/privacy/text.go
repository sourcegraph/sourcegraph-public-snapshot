package privacy

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Text is a helper type for tracking strings which may potentially
// be logged and/or contain user-private or org-private information.
//
// It acts as an "immutable" container; instead of allowing ad-hoc
// modification, methods return new Text values.
//
// NOTE: A String() method is deliberately omitted as accessing
// the underlying string data directly makes it easy to accidentally
// "launder" a Private Text into a Public Text. For common operations,
// using MapData or Combine should be sufficient. If you _really_ need
// direct access to the string data, use GetDataUnchecked.
type Text struct {
	data    string
	privacy Privacy
}

func (t Text) AsText() Text {
	return t
}

var _ AsText = Text{}

func NewText(data string, privacy Privacy) Text {
	return Text{data, privacy}
}

func NewTexts(data []string, privacy Privacy) []Text {
	out := make([]Text, 0, len(data))
	for _, d := range data {
		out = append(out, NewText(d, privacy))
	}
	return out
}

// GetDataUnchecked returns the underlying string data.
//
// This makes it possible to "launder" a Private Text into a Public Text.
//
// Avoid using this function if possible.
func (t Text) GetDataUnchecked() string {
	return t.data
}

func (t Text) Privacy() Privacy {
	return t.privacy
}

type TextJSON struct {
	Data    string `json:"data"`
	Privacy string `json:"privacy"`
}

func (t Text) MarshalJSON() ([]byte, error) {
	var privacy string
	switch t.Privacy() {
	case Private:
		privacy = "private"
	case Public:
		privacy = "public"
	default:
		privacy = "unknown"
	}
	return json.Marshal(&TextJSON{t.GetDataUnchecked(), privacy})
}

func (t *Text) UnmarshalJSON(data []byte) error {
	var tj TextJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return err
	}
	var privacy Privacy
	switch tj.Privacy {
	case "public":
		privacy = Public
	case "private":
		privacy = Private
	case "unknown":
		privacy = Unknown
	default:
		return errors.Newf("error: unrecognized privacy level %s", tj.Privacy)
	}
	*t = NewText(tj.Data, privacy)
	return nil
}

// MapText performs a textual transformation on the underlying data,
// maintaining the privacy level.
//
// It is a logic error to use this function to escape the string inside
// the Text value; use getDataUnchecked instead.
func (t Text) MapData(f func(string) string) Text {
	return NewText(f(t.data), t.privacy)
}

func (t Text) Combine(s Text, f func(string, string) string) Text {
	return NewText(f(t.data, s.data), t.privacy.Combine(s.privacy))
}

func (t Text) IncreasePrivacy(newPrivacy Privacy) Text {
	if t.privacy < newPrivacy {
		panic("error: Attempted to make value more public using IncreasePrivacy")
	}
	return NewText(t.data, newPrivacy)
}

// AsText represents types that are in 1-1 correspondence with Text.
//
// This interface is intended for reducing verbosity when logging Text values.
//
//   🙁 log.Text(key, privacy.Text(data)) // cast required for type Data privacy.Text
//   😐 log.Text(key, data.Value) // field projection required for type Data { Value privacy.Text }
//   😃 log.Text(key, data) // after implementing AsText for Data
type AsText interface {
	AsText() Text
}
