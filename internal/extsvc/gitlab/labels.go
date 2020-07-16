package gitlab

import "time"

type Label struct {
	ID          ID        `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	TextColor   string    `json:"text_color"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
