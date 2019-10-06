package graphqlbackend

import (
	"context"
)

type PartialComment interface {
	Author(context.Context) (*Actor, error)
	Body(context.Context) (string, error)
	BodyText(context.Context) (string, error)
	BodyHTML(context.Context) (string, error)
	CreatedAt(context.Context) (DateTime, error)
	UpdatedAt(context.Context) (DateTime, error)
}
