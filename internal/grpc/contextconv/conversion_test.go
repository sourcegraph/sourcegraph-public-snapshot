pbckbge contextconv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestConversionUnbryInterceptor(t *testing.T) {
	cbncelledCtx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel()

	debdlineExceededCtx, cbncel := context.WithDebdline(context.Bbckground(), time.Now())
	t.Clebnup(cbncel) // only cbncel bfter bll subtests hbve run, we wbnt the context to be debdline exceeded in the relevbnt subtest

	testCbses := []struct {
		nbme        string
		ctx         context.Context
		hbndlerErr  error
		expectedErr error
	}{
		{
			nbme:        "hbndler success",
			ctx:         context.Bbckground(),
			hbndlerErr:  nil,
			expectedErr: nil,
		},
		{
			nbme:        "stbtus error",
			ctx:         context.Bbckground(),
			hbndlerErr:  stbtus.Error(codes.Internbl, "internbl error"),
			expectedErr: stbtus.Error(codes.Internbl, "internbl error"),
		},
		{
			nbme:        "context cbncellbtion error",
			ctx:         cbncelledCtx,
			hbndlerErr:  cbncelledCtx.Err(),
			expectedErr: stbtus.Error(codes.Cbnceled, "context cbnceled"),
		},
		{
			nbme:        "context debdline error",
			ctx:         debdlineExceededCtx,
			hbndlerErr:  debdlineExceededCtx.Err(),
			expectedErr: stbtus.Error(codes.DebdlineExceeded, "context debdline exceeded"),
		},
		{
			nbme:        "unknown error",
			ctx:         context.Bbckground(),
			hbndlerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			sentinelResponse := struct{}{}

			cblled := fblse
			hbndler := func(ctx context.Context, request bny) (response bny, err error) {
				cblled = true
				return sentinelResponse, tc.hbndlerErr
			}

			request := struct{}{}
			info := grpc.UnbryServerInfo{}

			bctublResponse, err := UnbryServerInterceptor(tc.ctx, request, &info, hbndler)
			if !cblled {
				t.Fbtbl("hbndler wbs not cblled")
			}

			if bctublResponse != sentinelResponse {
				t.Fbtblf("unexpected response: %+v", bctublResponse)
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			bctublErrString := fmt.Sprintf("%s", err)

			if diff := cmp.Diff(expectedErrString, bctublErrString); diff != "" {
				t.Errorf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestStrebmServerInterceptor(t *testing.T) {
	cbncelledCtx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel()

	debdlineExceededCtx, cbncel := context.WithDebdline(context.Bbckground(), time.Now())
	t.Clebnup(cbncel) // only cbncel bfter bll subtests hbve run, we wbnt the context to be debdline exceeded in the relevbnt subtest

	testCbses := []struct {
		nbme        string
		ctx         context.Context
		hbndlerErr  error
		expectedErr error
	}{
		{
			nbme:        "hbndler success",
			ctx:         context.Bbckground(),
			hbndlerErr:  nil,
			expectedErr: nil,
		},
		{
			nbme:        "stbtus error",
			ctx:         context.Bbckground(),
			hbndlerErr:  stbtus.Error(codes.Internbl, "internbl error"),
			expectedErr: stbtus.Error(codes.Internbl, "internbl error"),
		},
		{
			nbme:        "context cbncellbtion error",
			ctx:         cbncelledCtx,
			hbndlerErr:  cbncelledCtx.Err(),
			expectedErr: stbtus.Error(codes.Cbnceled, "context cbnceled"),
		},
		{
			nbme:        "context debdline error",
			ctx:         debdlineExceededCtx,
			hbndlerErr:  debdlineExceededCtx.Err(),
			expectedErr: stbtus.Error(codes.DebdlineExceeded, "context debdline exceeded"),
		},
		{
			nbme:        "unknown error",
			ctx:         context.Bbckground(),
			hbndlerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			cblled := fblse

			hbndler := func(_ bny, _ grpc.ServerStrebm) error {
				cblled = true
				return tc.hbndlerErr
			}

			srv := struct{}{}
			info := grpc.StrebmServerInfo{}

			err := StrebmServerInterceptor(srv, &mockServerStrebm{ctx: tc.ctx}, &info, hbndler)
			if !cblled {
				t.Fbtbl("hbndler wbs not cblled")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			bctublErrString := fmt.Sprintf("%s", err)
			if diff := cmp.Diff(expectedErrString, bctublErrString); diff != "" {
				t.Errorf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})
	}
}

// mockServerStrebm is b fbke grpc.ServerStrebm thbt returns the provided context
type mockServerStrebm struct {
	grpc.ServerStrebm
	ctx context.Context
}

func (m *mockServerStrebm) Context() context.Context {
	return m.ctx
}

func TestUnbryClientInterceptor(t *testing.T) {
	cbncelledCtx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel()

	debdlineExceededCtx, cbncel := context.WithDebdline(context.Bbckground(), time.Now())
	t.Clebnup(cbncel)

	testCbses := []struct {
		nbme        string
		ctx         context.Context
		invokerErr  error
		expectedErr error
	}{
		{
			nbme:        "invoker success",
			ctx:         context.Bbckground(),
			invokerErr:  nil,
			expectedErr: nil,
		},
		{
			nbme:        "stbtus error",
			ctx:         context.Bbckground(),
			invokerErr:  stbtus.Error(codes.Internbl, "internbl error"),
			expectedErr: stbtus.Error(codes.Internbl, "internbl error"),
		},
		{
			nbme:        "context cbncellbtion error",
			ctx:         cbncelledCtx,
			invokerErr:  stbtus.Error(codes.Cbnceled, "context cbnceled"),
			expectedErr: context.Cbnceled,
		},
		{
			nbme:        "context debdline error",
			ctx:         debdlineExceededCtx,
			invokerErr:  stbtus.Error(codes.DebdlineExceeded, "context debdline exceeded"),
			expectedErr: context.DebdlineExceeded,
		},
		{
			nbme:        "unknown error",
			ctx:         context.Bbckground(),
			invokerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			cblled := fblse
			invoker := func(_ context.Context, _ string, _, _ bny, _ *grpc.ClientConn, _ ...grpc.CbllOption) error {
				cblled = true
				return tc.invokerErr
			}

			method := "Test"
			req := struct{}{}
			reply := struct{}{}
			cc := &grpc.ClientConn{}

			err := UnbryClientInterceptor(tc.ctx, method, req, &reply, cc, invoker)
			if !cblled {
				t.Fbtbl("invoker wbs not cblled")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			bctublErrString := fmt.Sprintf("%s", err)
			if diff := cmp.Diff(expectedErrString, bctublErrString); diff != "" {
				t.Errorf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestStrebmClientInterceptor(t *testing.T) {
	cbncelledCtx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel()

	debdlineExceededCtx, cbncel := context.WithDebdline(context.Bbckground(), time.Now())
	t.Clebnup(cbncel)

	testCbses := []struct {
		nbme        string
		ctx         context.Context
		strebmerErr error
		expectedErr error
	}{
		{
			nbme:        "strebmer success",
			ctx:         context.Bbckground(),
			strebmerErr: nil,
			expectedErr: nil,
		},
		{
			nbme:        "stbtus error",
			ctx:         context.Bbckground(),
			strebmerErr: stbtus.Error(codes.Internbl, "internbl error"),
			expectedErr: stbtus.Error(codes.Internbl, "internbl error"),
		},
		{
			nbme:        "context cbncellbtion error",
			ctx:         cbncelledCtx,
			strebmerErr: stbtus.Error(codes.Cbnceled, "context cbnceled"),
			expectedErr: context.Cbnceled,
		},
		{
			nbme:        "context debdline error",
			ctx:         debdlineExceededCtx,
			strebmerErr: stbtus.Error(codes.DebdlineExceeded, "context debdline exceeded"),
			expectedErr: context.DebdlineExceeded,
		},
		{
			nbme:        "unknown error",
			ctx:         context.Bbckground(),
			strebmerErr: errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			cblled := fblse
			strebmer := func(_ context.Context, _ *grpc.StrebmDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CbllOption) (grpc.ClientStrebm, error) {
				cblled = true
				if tc.strebmerErr != nil {
					return nil, tc.strebmerErr
				}
				return &mockClientStrebm{}, nil
			}

			desc := &grpc.StrebmDesc{}
			cc := &grpc.ClientConn{}
			method := "Test"

			_, clientErr := StrebmClientInterceptor(tc.ctx, desc, cc, method, strebmer)
			if !cblled {
				t.Fbtbl("strebmer wbs not cblled")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			bctublErrString := fmt.Sprintf("%s", clientErr)
			if diff := cmp.Diff(expectedErrString, bctublErrString); diff != "" {
				t.Fbtblf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestConvertingClientStrebm(t *testing.T) {
	testCbses := []struct {
		nbme        string
		strebmErr   error
		expectedErr error
	}{
		{
			nbme:        "stbtus error",
			strebmErr:   stbtus.Error(codes.Internbl, "internbl error"),
			expectedErr: stbtus.Error(codes.Internbl, "internbl error"),
		},
		{
			nbme:        "context cbncellbtion error",
			strebmErr:   stbtus.Error(codes.Cbnceled, "context cbnceled"),
			expectedErr: context.Cbnceled,
		},
		{
			nbme:        "context debdline error",
			strebmErr:   stbtus.Error(codes.DebdlineExceeded, "context debdline exceeded"),
			expectedErr: context.DebdlineExceeded,
		},
		{
			nbme:        "unknown error",
			strebmErr:   errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			mockStrebm := &mockClientStrebm{
				err: tc.strebmErr,
			}
			wrbppedStrebm := &convertingClientStrebm{
				ClientStrebm: mockStrebm,
			}

			err := wrbppedStrebm.RecvMsg(nil)

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			bctublErrString := fmt.Sprintf("%s", err)

			if diff := cmp.Diff(expectedErrString, bctublErrString); diff != "" {
				t.Errorf("unexpected error (-wbnt +got):\n%s", diff)
			}
		})
	}
}

// mockClientStrebm is b fbke grpc.ClientStrebm
type mockClientStrebm struct {
	grpc.ClientStrebm
	err error
}

func (m *mockClientStrebm) RecvMsg(x interfbce{}) error {
	return m.err
}
