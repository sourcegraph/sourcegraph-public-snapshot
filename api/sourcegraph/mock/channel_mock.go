package mock

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func (c *Channel_ListenClient) CloseSend() error {
	return nil
}

func (s *Channel_ListenClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (s *Channel_ListenClient) Trailer() metadata.MD {
	return nil
}

func (s *Channel_ListenClient) RecvMsg(m interface{}) error {
	return nil
}

func (s *Channel_ListenClient) SendMsg(m interface{}) error {
	return nil
}

func (c *Channel_ListenClient) Context() context.Context {
	return nil
}

func (c *Channel_ListenServer) Context() context.Context {
	return nil
}

func (s *Channel_ListenServer) RecvMsg(m interface{}) error {
	return nil
}

func (s *Channel_ListenServer) SendMsg(m interface{}) error {
	return nil
}

func (s *Channel_ListenServer) SendHeader(metadata.MD) error {
	return nil
}

func (s *Channel_ListenServer) SetTrailer(metadata.MD) {}
