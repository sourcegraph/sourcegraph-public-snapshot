package elastigo

import "encoding/json"

func NewHighlight() *HighlightDsl {
	return &HighlightDsl{}
}

type HighlightDsl struct {
	Settings  *HighlightEmbed           `-`
	TagSchema string                    `json:"tag_schema,omitempty"`
	Fields    map[string]HighlightEmbed `json:"fields,omitempty"`
}

func NewHighlightOpts() *HighlightEmbed {
	return &HighlightEmbed{}
}

type HighlightEmbed struct {
	BoundaryCharsVal   string    `json:"boundary_chars,omitempty"`
	BoundaryMaxScanVal int       `json:"boundary_max_scan,omitempty"`
	PreTags            []string  `json:"pre_tags,omitempty"`
	PostTags           []string  `json:"post_tags,omitempty"`
	FragmentSizeVal    int       `json:"fragment_size,omitempty"`
	NumOfFragmentsVal  int       `json:"number_of_fragments,omitempty"`
	HighlightQuery     *QueryDsl `json:"highlight_query,omitempty"`
	MatchedFieldsVal   []string  `json:"matched_fields,omitempty"`
	OrderVal           string    `json:"order,omitempty"`
	TypeVal            string    `json:"type,omitempty"`
}

// Custom marshalling
func (t *HighlightDsl) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	if t.Fields != nil {
		m["fields"] = t.Fields
	}

	if t.TagSchema != "" {
		m["tag_schema"] = t.TagSchema
	}

	if t.Settings == nil {
		return json.Marshal(m)
	}

	//This is terrible :(, could use structs package to avoid extra serialization.
	embed, err := json.Marshal(t.Settings)
	if err == nil {
		err = json.Unmarshal(embed, &m)
	}

	if err == nil {
		return json.Marshal(m)
	}

	return nil, err
}

func (h *HighlightDsl) AddField(name string, settings *HighlightEmbed) *HighlightDsl {
	if h.Fields == nil {
		h.Fields = make(map[string]HighlightEmbed)
	}

	if settings != nil {
		h.Fields[name] = *settings
	} else {
		h.Fields[name] = HighlightEmbed{}
	}

	return h
}

func (h *HighlightDsl) Schema(schema string) *HighlightDsl {
	h.TagSchema = schema
	return h
}

func (h *HighlightDsl) SetOptions(options *HighlightEmbed) *HighlightDsl {
	h.Settings = options
	return h
}

func (o *HighlightEmbed) BoundaryChars(chars string) *HighlightEmbed {
	o.BoundaryCharsVal = chars
	return o
}

func (o *HighlightEmbed) BoundaryMaxScan(max int) *HighlightEmbed {
	o.BoundaryMaxScanVal = max
	return o
}

func (he *HighlightEmbed) FragSize(size int) *HighlightEmbed {
	he.FragmentSizeVal = size
	return he
}

func (he *HighlightEmbed) NumFrags(numFrags int) *HighlightEmbed {
	he.NumOfFragmentsVal = numFrags
	return he
}

func (he *HighlightEmbed) MatchedFields(fields ...string) *HighlightEmbed {
	he.MatchedFieldsVal = fields
	return he
}

func (he *HighlightEmbed) Order(order string) *HighlightEmbed {
	he.OrderVal = order
	return he
}

func (he *HighlightEmbed) Tags(pre string, post string) *HighlightEmbed {
	if he == nil {
		he = &HighlightEmbed{}
	}

	if he.PreTags == nil {
		he.PreTags = []string{pre}
	} else {
		he.PreTags = append(he.PreTags, pre)
	}

	if he.PostTags == nil {
		he.PostTags = []string{post}
	} else {
		he.PostTags = append(he.PostTags, post)
	}

	return he
}

func (he *HighlightEmbed) Type(highlightType string) *HighlightEmbed {
	he.TypeVal = highlightType
	return he
}
