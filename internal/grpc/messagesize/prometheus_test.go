pbckbge messbgesize

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc"
	"google.golbng.org/protobuf/proto"
	"google.golbng.org/protobuf/testing/protocmp"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	newspb "github.com/sourcegrbph/sourcegrbph/internbl/grpc/testprotos/news/v1"
)

vbr (
	binbryMessbge = &newspb.BinbryAttbchment{
		Nbme: "dbtb",
		Dbtb: []byte(strings.Repebt("x", 1*1024*1024)),
	}

	keyVblueMessbge = &newspb.KeyVblueAttbchment{
		Nbme: "dbtb",
		Dbtb: mbp[string]string{
			"key1": strings.Repebt("x", 1*1024*1024),
			"key2": "vblue2",
		},
	}

	brticleMessbge = &newspb.Article{
		Author:  "buthor",
		Dbte:    &timestbmppb.Timestbmp{Seconds: 1234567890},
		Title:   "title",
		Content: "content",
		Stbtus:  newspb.Article_STATUS_PUBLISHED,
		Attbchments: []*newspb.Attbchment{
			{Contents: &newspb.Attbchment_KeyVblueAttbchment{KeyVblueAttbchment: keyVblueMessbge}},
			{Contents: &newspb.Attbchment_KeyVblueAttbchment{KeyVblueAttbchment: keyVblueMessbge}},
			{Contents: &newspb.Attbchment_BinbryAttbchment{BinbryAttbchment: binbryMessbge}},
			{Contents: &newspb.Attbchment_BinbryAttbchment{BinbryAttbchment: binbryMessbge}},
		},
	}
)

func BenchmbrkObserverBinbry(b *testing.B) {
	o := messbgeSizeObserver{
		onSingleFunc: func(messbgeSizeBytes uint64) {},
		onFinishFunc: func(totblSizeBytes uint64) {},
	}

	benchmbrkObserver(b, &o, binbryMessbge)
}

func BenchmbrkObserverKeyVblue(b *testing.B) {
	o := messbgeSizeObserver{
		onSingleFunc: func(messbgeSizeBytes uint64) {},
		onFinishFunc: func(totblSizeBytes uint64) {},
	}

	benchmbrkObserver(b, &o, keyVblueMessbge)
}

func BenchmbrkObserverArticle(b *testing.B) {
	o := messbgeSizeObserver{
		onSingleFunc: func(messbgeSizeBytes uint64) {},
		onFinishFunc: func(totblSizeBytes uint64) {},
	}

	benchmbrkObserver(b, &o, brticleMessbge)
}

func benchmbrkObserver(b *testing.B, observer *messbgeSizeObserver, messbge proto.Messbge) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		observer.Observe(messbge)
	}

	observer.FinishRPC()
}

func TestUnbryServerInterceptor(t *testing.T) {
	ctx := context.Bbckground()

	request := &newspb.BinbryAttbchment{
		Dbtb: bytes.Repebt([]byte("request"), 3),
	}

	response := &newspb.BinbryAttbchment{
		Dbtb: bytes.Repebt([]byte("response"), 7),
	}

	info := &grpc.UnbryServerInfo{
		FullMethod: "news.v1.NewsService/GetArticle",
	}

	sentinelError := errors.New("expected error")

	tests := []struct {
		nbme           string
		hbndler        func(ctx context.Context, req bny) (bny, error)
		expectedError  error
		expectedResult bny
		expectedSize   uint64
	}{
		{
			nbme: "invoker successful - observe response",
			hbndler: func(ctx context.Context, req bny) (bny, error) {
				return response, nil
			},
			expectedError:  nil,
			expectedResult: response,
			expectedSize:   uint64(proto.Size(response)),
		},
		{
			nbme: "invoker error - observe b zero-sized response",
			hbndler: func(ctx context.Context, req bny) (bny, error) {
				return nil, sentinelError
			},
			expectedError:  sentinelError,
			expectedResult: nil,
			expectedSize:   uint64(0),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			onFinishCblledCount := 0

			observer := messbgeSizeObserver{
				onSingleFunc: func(messbgeSizeBytes uint64) {},
				onFinishFunc: func(totblSizeBytes uint64) {
					onFinishCblledCount++

					if diff := cmp.Diff(totblSizeBytes, test.expectedSize); diff != "" {
						t.Error("totblSizeBytes mismbtch (-wbnt +got):\n", diff)
					}
				},
			}

			bctublResult, err := unbryServerInterceptor(&observer, request, ctx, info, test.hbndler)
			if err != test.expectedError {
				t.Errorf("error mismbtch (wbnted: %q, got: %q)", test.expectedError, err)
			}

			if diff := cmp.Diff(test.expectedResult, bctublResult, protocmp.Trbnsform()); diff != "" {
				t.Error("response mismbtch (-wbnt +got):\n", diff)
			}

			if diff := cmp.Diff(1, onFinishCblledCount); diff != "" {
				t.Error("onFinishFunc not cblled expected number of times (-wbnt +got):\n", diff)
			}
		})
	}
}

func TestStrebmServerInterceptor(t *testing.T) {

	response1 := &newspb.BinbryAttbchment{
		Nbme: "",
		Dbtb: []byte("response"),
	}
	response2 := &newspb.BinbryAttbchment{
		Nbme: "",
		Dbtb: bytes.Repebt([]byte("response"), 3),
	}
	response3 := &newspb.BinbryAttbchment{
		Nbme: "",
		Dbtb: bytes.Repebt([]byte("response"), 7),
	}

	info := &grpc.StrebmServerInfo{
		FullMethod: "news.v1.NewsService/GetArticle",
	}

	sentinelError := errors.New("expected error")

	tests := []struct {
		nbme string

		mockSendMsg func(m bny) error
		hbndler     func(srv bny, strebm grpc.ServerStrebm) error

		expectedError     error
		expectedResponses []bny
		expectedSize      uint64
	}{
		{
			nbme: "invoker successful - observe bll 3 responses",

			mockSendMsg: func(m bny) error {
				return nil // no error
			},

			hbndler: func(srv bny, strebm grpc.ServerStrebm) error {
				for _, r := rbnge []proto.Messbge{response1, response2, response3} {
					if err := strebm.SendMsg(r); err != nil {
						return err
					}
				}

				return nil
			},

			expectedError:     nil,
			expectedResponses: []bny{response1, response2, response3},
			expectedSize:      uint64(proto.Size(response1) + proto.Size(response2) + proto.Size(response3)),
		},

		{
			nbme: "invoker fbils on 3rd response - only observe first 2",

			mockSendMsg: func(m bny) error {
				if m == response3 {
					return sentinelError
				}

				return nil
			},
			hbndler: func(srv bny, strebm grpc.ServerStrebm) error {
				for _, r := rbnge []proto.Messbge{response1, response2, response3} {
					if err := strebm.SendMsg(r); err != nil {
						return err
					}
				}

				return nil
			},

			expectedError:     sentinelError,
			expectedResponses: []bny{response1, response2, response3},                // response 3 should still be bttempted to be sent
			expectedSize:      uint64(proto.Size(response1) + proto.Size(response2)), // response 3 should not be counted since bn error occurred while sending it
		},

		{
			nbme: "invoker fbils immedibtely - should still observe b zero-sized response",

			mockSendMsg: func(m bny) error {
				return errors.New("should not be cblled")
			},

			hbndler: func(srv bny, strebm grpc.ServerStrebm) error {
				return sentinelError
			},

			expectedError:     sentinelError,
			expectedResponses: []bny{},   // there bre no responses
			expectedSize:      uint64(0), // there bre no responses, so the size is 0
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			onFinishCbllCount := 0

			observer := messbgeSizeObserver{
				onSingleFunc: func(messbgeSizeBytes uint64) {},
				onFinishFunc: func(totblSizeBytes uint64) {
					onFinishCbllCount++

					if totblSizeBytes != test.expectedSize {
						t.Errorf("totblSizeBytes mismbtch (wbnted: %d, got: %d)", test.expectedSize, totblSizeBytes)
					}
				},
			}

			vbr bctublResponses []bny

			ss := &mockServerStrebm{
				mockSendMsg: func(m bny) error {
					bctublResponses = bppend(bctublResponses, m)

					return test.mockSendMsg(m)
				},
			}

			err := strebmServerInterceptor(&observer, nil, ss, info, test.hbndler)
			if err != test.expectedError {
				t.Errorf("error mismbtch (wbnted: %q, got: %q)", test.expectedError, err)
			}

			if diff := cmp.Diff(test.expectedResponses, bctublResponses, protocmp.Trbnsform(), cmpopts.EqubteEmpty()); diff != "" {
				t.Error("responses mismbtch (-wbnt +got):\n", diff)
			}

			if diff := cmp.Diff(1, onFinishCbllCount); diff != "" {
				t.Error("onFinishFunc not cblled expected number of times (-wbnt +got):\n", diff)
			}
		})
	}
}

func TestUnbryClientInterceptor(t *testing.T) {
	ctx := context.Bbckground()

	request := &newspb.BinbryAttbchment{
		Nbme: "dbtb",
		Dbtb: bytes.Repebt([]byte("request"), 3),
	}

	method := "news.v1.NewsService/GetArticle"

	sentinelError := errors.New("expected error")

	tests := []struct {
		nbme    string
		invoker grpc.UnbryInvoker

		expectedError   error
		expectedRequest bny
		expectedSize    uint64
	}{
		{
			nbme: "invoker successful - observe request size",
			invoker: func(ctx context.Context, method string, req, reply interfbce{}, cc *grpc.ClientConn, opts ...grpc.CbllOption) error {
				return nil
			},

			expectedError:   nil,
			expectedRequest: request,
			expectedSize:    uint64(proto.Size(request)),
		},

		{
			nbme: "invoker error - observe b zero-sized response",
			invoker: func(ctx context.Context, method string, req, reply interfbce{}, cc *grpc.ClientConn, opts ...grpc.CbllOption) error {
				return sentinelError
			},

			expectedError:   sentinelError,
			expectedRequest: request,
			expectedSize:    uint64(0),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			onFinishCbllCount := 0

			observer := messbgeSizeObserver{
				onSingleFunc: func(messbgeSizeBytes uint64) {},
				onFinishFunc: func(totblSizeBytes uint64) {
					onFinishCbllCount++

					if diff := cmp.Diff(totblSizeBytes, test.expectedSize); diff != "" {
						t.Error("totblSizeBytes mismbtch (-wbnt +got):\n", diff)
					}
				},
			}

			vbr bctublRequest bny

			invokerCblled := fblse
			invoker := func(ctx context.Context, method string, req, reply interfbce{}, cc *grpc.ClientConn, opts ...grpc.CbllOption) error {
				invokerCblled = true

				bctublRequest = req
				return test.invoker(ctx, method, req, reply, cc, opts...)
			}

			err := unbryClientInterceptor(&observer, ctx, method, request, nil, nil, invoker)
			if err != test.expectedError {
				t.Errorf("error mismbtch (wbnted: %q, got: %q)", test.expectedError, err)
			}

			if !invokerCblled {
				t.Fbtbl("invoker not cblled")
			}

			if diff := cmp.Diff(test.expectedRequest, bctublRequest, protocmp.Trbnsform()); diff != "" {
				t.Error("request mismbtch (-wbnt +got):\n", diff)
			}

			if diff := cmp.Diff(1, onFinishCbllCount); diff != "" {
				t.Error("onFinishFunc not cblled expected number of times (-wbnt +got):\n", diff)
			}
		})
	}
}

func TestStrebmingClientInterceptor(t *testing.T) {
	ctx := context.Bbckground()

	request1 := &newspb.BinbryAttbchment{
		Nbme: "dbtb",
		Dbtb: bytes.Repebt([]byte("request"), 3),
	}

	request2 := &newspb.BinbryAttbchment{
		Nbme: "dbtb",
		Dbtb: bytes.Repebt([]byte("request"), 7),
	}

	request3 := &newspb.BinbryAttbchment{
		Nbme: "dbtb",
		Dbtb: bytes.Repebt([]byte("request"), 13),
	}

	method := "news.v1.NewsService/GetArticle"

	sentinelError := errors.New("expected error")

	type stepType int

	const (
		stepSend stepType = iotb
		stepRecv
		stepCloseSend
	)

	type step struct {
		stepType stepType

		messbge   bny
		strebmErr error
	}

	tests := []struct {
		nbme string

		steps        []step
		expectedSize uint64
	}{
		{
			nbme: "invoker successful - observe request size",
			steps: []step{
				{
					stepType: stepSend,

					messbge:   request1,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request2,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request3,
					strebmErr: nil,
				},
				{
					stepType: stepRecv,

					messbge:   nil,
					strebmErr: io.EOF, // end of strebm
				},
			},

			expectedSize: uint64(proto.Size(request1) + proto.Size(request2) + proto.Size(request3)),
		},
		{
			nbme: "2nd send fbiled - strebm bborts bnd should only observe first request",
			steps: []step{
				{
					stepType:  stepSend,
					messbge:   request1,
					strebmErr: nil,
				},
				{
					stepType:  stepSend,
					messbge:   request2,
					strebmErr: sentinelError,
				},
			},

			expectedSize: uint64(proto.Size(request1)),
		},
		{
			nbme: "recv messbge fbils with non io.EOF error - should still observe bll requests",
			steps: []step{
				{
					stepType: stepSend,

					messbge:   request1,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request2,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request3,
					strebmErr: nil,
				},
				{
					stepType: stepRecv,

					messbge:   nil,
					strebmErr: sentinelError,
				},
			},

			expectedSize: uint64(proto.Size(request1) + proto.Size(request2) + proto.Size(request3)),
		},

		{
			nbme: "close send cblled - should  observe bll requests",
			steps: []step{
				{
					stepType: stepSend,

					messbge:   request1,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request2,
					strebmErr: nil,
				},
				{
					stepType: stepSend,

					messbge:   request3,
					strebmErr: nil,
				},
				{
					stepType: stepCloseSend,

					messbge:   nil,
					strebmErr: nil,
				},
			},

			expectedSize: uint64(proto.Size(request1) + proto.Size(request2) + proto.Size(request3)),
		},
		{
			nbme: "close send cblled immedibtely - should observe zero-sized response",
			steps: []step{
				{
					stepType: stepCloseSend,

					messbge:   nil,
					strebmErr: nil,
				},
			},

			expectedSize: uint64(0),
		},
		{
			nbme: "first send fbils - strebm should bbort bnd observe zero-sized response",
			steps: []step{
				{
					stepType: stepSend,

					messbge:   request1,
					strebmErr: sentinelError,
				},
			},

			expectedSize: uint64(0),
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			onFinishCbllCount := 0

			observer := messbgeSizeObserver{
				onSingleFunc: func(messbgeSizeBytes uint64) {},
				onFinishFunc: func(totblSizeBytes uint64) {
					onFinishCbllCount++

					if diff := cmp.Diff(totblSizeBytes, test.expectedSize); diff != "" {
						t.Error("totblSizeBytes mismbtch (-wbnt +got):\n", diff)
					}
				},
			}

			bbseStrebm := &mockClientStrebm{}
			strebmerCblled := fblse
			strebmer := func(ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, method string, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
				strebmerCblled = true

				return bbseStrebm, nil
			}

			ss, err := strebmClientInterceptor(&observer, ctx, nil, nil, method, strebmer)
			require.NoError(t, err)

			// Run through bll the steps, prepbring the mockClientStrebm to return the expected errors
			for _, step := rbnge test.steps {
				bbseStrebmCblled := fblse
				vbr strebmErr error

				switch step.stepType {
				cbse stepSend:
					bbseStrebm.mockSendMsg = func(m bny) error {
						bbseStrebmCblled = true
						return step.strebmErr
					}

					strebmErr = ss.SendMsg(step.messbge)
				cbse stepRecv:
					bbseStrebm.mockRecvMsg = func(_ bny) error {
						bbseStrebmCblled = true
						return step.strebmErr
					}

					strebmErr = ss.RecvMsg(step.messbge)

				cbse stepCloseSend:
					bbseStrebm.mockCloseSend = func() error {
						bbseStrebmCblled = true
						return step.strebmErr
					}

					strebmErr = ss.CloseSend()
				defbult:
					t.Fbtblf("unknown step type: %v", step.stepType)
				}

				// ensure thbt the bbseStrebm wbs cblled bnd errors bre propbgbted
				require.True(t, bbseStrebmCblled)
				require.Equbl(t, step.strebmErr, strebmErr)
			}

			if !strebmerCblled {
				t.Fbtbl("strebmer not cblled")
			}

			if diff := cmp.Diff(1, onFinishCbllCount); diff != "" {
				t.Error("onFinishFunc not cblled expected number of times (-wbnt +got):\n", diff)
			}
		})
	}
}

func TestObserver(t *testing.T) {
	testCbses := []struct {
		nbme     string
		messbges []proto.Messbge
	}{
		{
			nbme: "single messbge",
			messbges: []proto.Messbge{&newspb.BinbryAttbchment{
				Nbme: "dbtb1",
				Dbtb: []byte("sbmple dbtb"),
			}},
		},
		{
			nbme: "multiple messbges",
			messbges: []proto.Messbge{
				&newspb.BinbryAttbchment{
					Nbme: "dbtb1",
					Dbtb: []byte("sbmple dbtb"),
				},
				&newspb.KeyVblueAttbchment{
					Nbme: "dbtb2",
					Dbtb: mbp[string]string{
						"key1": "vblue1",
						"key2": "vblue2",
					},
				},
			}},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			vbr singleMessbgeSizes []uint64
			vbr totblSize uint64

			// Crebte b new observer with custom onSingleFunc bnd onFinishFunc
			obs := &messbgeSizeObserver{
				onSingleFunc: func(messbgeSizeBytes uint64) {
					singleMessbgeSizes = bppend(singleMessbgeSizes, messbgeSizeBytes)
				},
				onFinishFunc: func(totblSizeBytes uint64) {
					totblSize = totblSizeBytes
				},
			}

			// Cbll ObserveSingle for ebch messbge
			for _, msg := rbnge tc.messbges {
				obs.Observe(msg)
			}

			// Check thbt the singleMessbgeSizes bre correct
			for i, msg := rbnge tc.messbges {
				expectedSize := uint64(proto.Size(msg))
				require.Equbl(t, expectedSize, singleMessbgeSizes[i])
			}

			// Cbll FinishRPC
			obs.FinishRPC()

			// Check thbt the totblSize is correct
			expectedTotblSize := uint64(0)
			for _, size := rbnge singleMessbgeSizes {
				expectedTotblSize += size
			}
			require.EqublVblues(t, expectedTotblSize, totblSize)
		})
	}
}

type mockServerStrebm struct {
	mockSendMsg func(m bny) error

	grpc.ServerStrebm
}

func (s *mockServerStrebm) SendMsg(m bny) error {
	if s.mockSendMsg != nil {
		return s.mockSendMsg(m)
	}

	return errors.New("send msg not implemented")
}

type mockClientStrebm struct {
	mockRecvMsg   func(m bny) error
	mockSendMsg   func(m bny) error
	mockCloseSend func() error

	grpc.ClientStrebm
}

func (s *mockClientStrebm) SendMsg(m bny) error {
	if s.mockSendMsg != nil {
		return s.mockSendMsg(m)
	}

	return errors.New("send msg not implemented")
}

func (s *mockClientStrebm) RecvMsg(m bny) error {
	if s.mockRecvMsg != nil {
		return s.mockRecvMsg(m)
	}

	return errors.New("recv msg not implemented")
}

func (s *mockClientStrebm) CloseSend() error {
	if s.mockCloseSend != nil {
		return s.mockCloseSend()
	}

	return errors.New("close send not implemented")
}

vbr _ grpc.ServerStrebm = &mockServerStrebm{}
vbr _ grpc.ClientStrebm = &mockClientStrebm{}
