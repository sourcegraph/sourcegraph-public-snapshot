package graphqlbackend

import "context"

type updatable interface {
	ViewerCanUpdate(context.Context) (bool, error)
}
