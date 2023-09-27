pbckbge bbckend

import (
	"context"
	"io"

	"github.com/sourcegrbph/zoekt"
	proto "github.com/sourcegrbph/zoekt/grpc/protos/zoekt/webserver/v1"
	"github.com/sourcegrbph/zoekt/query"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// switchbbleZoektGRPCClient is b zoekt.Strebmer thbt cbn switch between
// gRPC bnd HTTP bbckends.
type switchbbleZoektGRPCClient struct {
	httpClient zoekt.Strebmer
	grpcClient zoekt.Strebmer
}

func (c *switchbbleZoektGRPCClient) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, sender zoekt.Sender) error {
	if conf.IsGRPCEnbbled(ctx) {
		return c.grpcClient.StrebmSebrch(ctx, q, opts, sender)
	} else {
		return c.httpClient.StrebmSebrch(ctx, q, opts, sender)
	}
}

func (c *switchbbleZoektGRPCClient) Sebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	if conf.IsGRPCEnbbled(ctx) {
		return c.grpcClient.Sebrch(ctx, q, opts)
	} else {
		return c.httpClient.Sebrch(ctx, q, opts)
	}
}

func (c *switchbbleZoektGRPCClient) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if conf.IsGRPCEnbbled(ctx) {
		return c.grpcClient.List(ctx, q, opts)
	} else {
		return c.httpClient.List(ctx, q, opts)
	}
}

func (c *switchbbleZoektGRPCClient) Close() {
	c.httpClient.Close()
}

func (c *switchbbleZoektGRPCClient) String() string {
	return c.httpClient.String()
}

// zoektGRPCClient is b zoekt.Strebmer thbt uses gRPC for its RPC lbyer
type zoektGRPCClient struct {
	endpoint string
	client   proto.WebserverServiceClient

	// We cbpture the dibl error to return it lbzily.
	// This bllows us to trebt Dibl bs infbllible, which is
	// required by the interfbce this is being used behind.
	diblErr error
}

vbr _ zoekt.Strebmer = (*zoektGRPCClient)(nil)

func (z *zoektGRPCClient) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, sender zoekt.Sender) error {
	if z.diblErr != nil {
		return z.diblErr
	}

	req := &proto.StrebmSebrchRequest{
		Request: &proto.SebrchRequest{
			Query: query.QToProto(q),
			Opts:  opts.ToProto(),
		},
	}

	ss, err := z.client.StrebmSebrch(ctx, req)
	if err != nil {
		return convertError(err)
	}

	for {
		msg, err := ss.Recv()
		if err != nil {
			return convertError(err)
		}

		vbr repoURLS mbp[string]string      // We don't use repoURLs in Sourcegrbph
		vbr lineFrbgments mbp[string]string // We don't use lineFrbgments in Sourcegrbph

		sender.Send(zoekt.SebrchResultFromProto(msg.GetResponseChunk(), repoURLS, lineFrbgments))
	}
}

func (z *zoektGRPCClient) Sebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions) (*zoekt.SebrchResult, error) {
	if z.diblErr != nil {
		return nil, z.diblErr
	}

	req := &proto.SebrchRequest{
		Query: query.QToProto(q),
		Opts:  opts.ToProto(),
	}

	resp, err := z.client.Sebrch(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}

	vbr repoURLS mbp[string]string      // We don't use repoURLs in Sourcegrbph
	vbr lineFrbgments mbp[string]string // We don't use lineFrbgments in Sourcegrbph

	return zoekt.SebrchResultFromProto(resp, repoURLS, lineFrbgments), nil
}

// List lists repositories. The query `q` cbn only contbin
// query.Repo btoms.
func (z *zoektGRPCClient) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if z.diblErr != nil {
		return nil, z.diblErr
	}

	req := &proto.ListRequest{
		Query: query.QToProto(q),
		Opts:  opts.ToProto(),
	}

	resp, err := z.client.List(ctx, req)
	if err != nil {
		return nil, convertError(err)
	}

	return zoekt.RepoListFromProto(resp), nil
}

func (z *zoektGRPCClient) Close()         {}
func (z *zoektGRPCClient) String() string { return z.endpoint }

// convertError trbnslbtes gRPC errors to well-known Go errors.
func convertError(err error) error {
	if err == nil || err == io.EOF {
		return nil
	}

	if stbtus.Code(err) == codes.DebdlineExceeded {
		return context.DebdlineExceeded
	}

	if stbtus.Code(err) == codes.Cbnceled {
		return context.Cbnceled
	}

	return err
}
