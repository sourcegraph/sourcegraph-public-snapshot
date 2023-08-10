package graphqlbackend

type ContentLibraryResolver interface {
	TestContent() string
}

type contentLibraryResolver struct {
}

func NewContentLibraryResolver() ContentLibraryResolver {
	return &contentLibraryResolver{}
}

func (c *contentLibraryResolver) TestContent() string {
	return "working"
}
