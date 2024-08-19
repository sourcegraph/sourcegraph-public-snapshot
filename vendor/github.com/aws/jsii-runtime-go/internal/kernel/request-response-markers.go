package kernel

// kernelRequester allows creating a union of kernelRequester and kernelResponder
// types by defining private method implemented by a private custom type, which
// is embedded in all relevant types.
type kernelRequester interface {
	// isRequest is a marker method that is intended to ensure no external type
	// can implement this interface.
	isRequest() kernelBrand
}

// kernelRequest is the standard implementation for kernelRequester.
type kernelRequest struct {
	API string `json:"api"`
}

func (r kernelRequest) isRequest() kernelBrand {
	return kernelBrand{}
}

// kernelResponder allows creating a union of kernelResponder and kernelRequester
// types by defining private method implemented by a private custom type, which
// is embedded in all relevant types.
type kernelResponder interface {
	// isResponse is a marker method that is intended to ensure no external type
	// can implement this interface.
	isResponse() kernelBrand
}

// kernelResponse is a 0-width marker struc tembedded to make another type be
// usable as a kernelResponder.
type kernelResponse struct{}

func (r kernelResponse) isResponse() kernelBrand {
	return kernelBrand{}
}

// kernelBrand is a private type used to ensure the kernelRequester and
// kernelResponder cannot be implemented outside of this package.
type kernelBrand struct{}
