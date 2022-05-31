package privacy

// Text is a helper type for tracking strings which may potentially
// be logged and/or contain user-private or org-private information.
//
// It acts as an "immutable" container; instead of allowing ad-hoc
// modification, methods return new Text values.
type Text struct {
	data    string
	privacy Privacy
}

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
