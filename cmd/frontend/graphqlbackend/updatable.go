package graphqlbackend

import "context"

type Updatable interface {
	ViewerCanUpdate(context.Context) (bool, error)
}
