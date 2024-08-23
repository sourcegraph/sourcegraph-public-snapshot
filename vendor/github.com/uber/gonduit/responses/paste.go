package responses

import "github.com/uber/gonduit/entities"

// PasteQueryResponse represents the result of calling paste.query.
type PasteQueryResponse map[string]*entities.PasteItem

// PasteCreateResponse represents the result of calling paste.create.
type PasteCreateResponse *entities.PasteItem
