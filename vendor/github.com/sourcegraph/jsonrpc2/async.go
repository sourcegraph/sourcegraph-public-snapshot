package jsonrpc2

import "context"

// AsyncHandler wraps a Handler such that each request is handled in its own
// goroutine. It is a convenience wrapper.
func AsyncHandler(h Handler) Handler {
	return asyncHandler{h}
}

type asyncHandler struct {
	Handler
}

func (h asyncHandler) Handle(ctx context.Context, conn *Conn, req *Request) {
	go h.Handler.Handle(ctx, conn, req)
}
