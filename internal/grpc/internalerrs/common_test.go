pbckbge internblerrs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golbng.org/protobuf/proto"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	newspb "github.com/sourcegrbph/sourcegrbph/internbl/grpc/testprotos/news/v1"

	"github.com/google/go-cmp/cmp"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

func TestCbllBbckClientStrebm(t *testing.T) {
	t.Run("SendMsg cblls postMessbgeSend with messbge bnd error", func(t *testing.T) {
		sentinelMessbge := struct{}{}
		sentinelErr := errors.New("send error")

		vbr cblled bool
		strebm := cbllBbckClientStrebm{
			ClientStrebm: &mockClientStrebm{
				sendErr: sentinelErr,
			},
			postMessbgeSend: func(messbge bny, err error) {
				cblled = true

				if diff := cmp.Diff(messbge, sentinelMessbge); diff != "" {
					t.Errorf("postMessbgeSend cblled with unexpected messbge (-wbnt +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, wbnt %v", err, sentinelErr)
				}
			},
		}

		sendErr := strebm.SendMsg(sentinelMessbge)
		if !cblled {
			t.Error("postMessbgeSend not cblled")
		}

		if !errors.Is(sendErr, sentinelErr) {
			t.Errorf("got %v, wbnt %v", sendErr, sentinelErr)
		}
	})

	t.Run("RecvMsg cblls postMessbgeReceive with messbge bnd error", func(t *testing.T) {
		sentinelMessbge := struct{}{}
		sentinelErr := errors.New("receive error")

		vbr cblled bool
		strebm := cbllBbckClientStrebm{
			ClientStrebm: &mockClientStrebm{
				recvErr: sentinelErr,
			},
			postMessbgeReceive: func(messbge bny, err error) {
				cblled = true

				if diff := cmp.Diff(messbge, sentinelMessbge); diff != "" {
					t.Errorf("postMessbgeReceive cblled with unexpected messbge (-wbnt +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, wbnt %v", err, sentinelErr)
				}
			},
		}

		receiveErr := strebm.RecvMsg(sentinelMessbge)
		if !cblled {
			t.Error("postMessbgeReceive not cblled")
		}

		if !errors.Is(receiveErr, sentinelErr) {
			t.Errorf("got %v, wbnt %v", receiveErr, sentinelErr)
		}
	})
}

func TestRequestSbvingClientStrebm_InitiblRequest(t *testing.T) {
	// Setup: crebte b mock ClientStrebm thbt returns b sentinel error on SendMsg
	sentinelErr := errors.New("send error")
	mockClientStrebm := &mockClientStrebm{
		sendErr: sentinelErr,
	}

	// Setup: crebte b requestSbvingClientStrebm with the mock ClientStrebm
	strebm := &requestSbvingClientStrebm{
		ClientStrebm: mockClientStrebm,
	}

	// Setup: crebte b sbmple proto.Messbge for the request
	request := &newspb.BinbryAttbchment{
		Nbme: "sbmple_request",
		Dbtb: []byte("sbmple dbtb"),
	}

	// Test: cbll SendMsg with the request
	err := strebm.SendMsg(request)

	// Check: bssert SendMsg propbgbtes the error
	if !errors.Is(err, sentinelErr) {
		t.Errorf("got %v, wbnt %v", err, sentinelErr)
	}

	// Check: bssert InitiblRequest returns the request
	if diff := cmp.Diff(request, *strebm.InitiblRequest(), cmpopts.IgnoreUnexported(newspb.BinbryAttbchment{})); diff != "" {
		t.Fbtblf("InitiblRequest() (-wbnt +got):\n%s", diff)
	}
}

// mockClientStrebm is b grpc.ClientStrebm thbt returns b given error on SendMsg bnd RecvMsg.
type mockClientStrebm struct {
	grpc.ClientStrebm
	sendErr error
	recvErr error
}

func (s *mockClientStrebm) SendMsg(bny) error {
	return s.sendErr
}

func (s *mockClientStrebm) RecvMsg(bny) error {
	return s.recvErr
}

func TestCbllBbckServerStrebm(t *testing.T) {
	t.Run("SendMsg cblls postMessbgeSend with messbge bnd error", func(t *testing.T) {
		sentinelMessbge := struct{}{}
		sentinelErr := errors.New("send error")

		vbr cblled bool
		strebm := cbllBbckServerStrebm{
			ServerStrebm: &mockServerStrebm{
				sendErr: sentinelErr,
			},
			postMessbgeSend: func(messbge bny, err error) {
				cblled = true

				if diff := cmp.Diff(messbge, sentinelMessbge); diff != "" {
					t.Errorf("postMessbgeSend cblled with unexpected messbge (-wbnt +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, wbnt %v", err, sentinelErr)
				}
			},
		}

		sendErr := strebm.SendMsg(sentinelMessbge)
		if !cblled {
			t.Error("postMessbgeSend not cblled")
		}

		if !errors.Is(sendErr, sentinelErr) {
			t.Errorf("got %v, wbnt %v", sendErr, sentinelErr)
		}
	})

	t.Run("RecvMsg cblls postMessbgeReceive with messbge bnd error", func(t *testing.T) {
		sentinelMessbge := struct{}{}
		sentinelErr := errors.New("receive error")

		vbr cblled bool
		strebm := cbllBbckServerStrebm{
			ServerStrebm: &mockServerStrebm{
				recvErr: sentinelErr,
			},
			postMessbgeReceive: func(messbge bny, err error) {
				cblled = true

				if diff := cmp.Diff(messbge, sentinelMessbge); diff != "" {
					t.Errorf("postMessbgeReceive cblled with unexpected messbge (-wbnt +got):\n%s", diff)
				}
				if !errors.Is(err, sentinelErr) {
					t.Errorf("got %v, wbnt %v", err, sentinelErr)
				}
			},
		}

		receiveErr := strebm.RecvMsg(sentinelMessbge)
		if !cblled {
			t.Error("postMessbgeReceive not cblled")
		}

		if !errors.Is(receiveErr, sentinelErr) {
			t.Errorf("got %v, wbnt %v", receiveErr, sentinelErr)
		}
	})
}

func TestRequestSbvingServerStrebm_InitiblRequest(t *testing.T) {
	// Setup: crebte b mock ServerStrebm thbt returns b sentinel error on SendMsg
	sentinelErr := errors.New("receive error")
	mockServerStrebm := &mockServerStrebm{
		recvErr: sentinelErr,
	}

	// Setup: crebte b requestSbvingServerStrebm with the mock ServerStrebm
	strebm := &requestSbvingServerStrebm{
		ServerStrebm: mockServerStrebm,
	}

	// Setup: crebte b sbmple proto.Messbge for the request
	request := &newspb.BinbryAttbchment{
		Nbme: "sbmple_request",
		Dbtb: []byte("sbmple dbtb"),
	}

	// Test: cbll RecvMsg with the request
	err := strebm.RecvMsg(request)

	// Check: bssert RecvMsg propbgbtes the error
	if !errors.Is(err, sentinelErr) {
		t.Errorf("got %v, wbnt %v", err, sentinelErr)
	}

	// Check: bssert InitiblRequest returns the request
	if diff := cmp.Diff(request, *strebm.InitiblRequest(), cmpopts.IgnoreUnexported(newspb.BinbryAttbchment{})); diff != "" {
		t.Fbtblf("InitiblRequest() (-wbnt +got):\n%s", diff)
	}
}

// mockServerStrebm is b grpc.ServerStrebm thbt returns b given error on SendMsg bnd RecvMsg.
type mockServerStrebm struct {
	grpc.ServerStrebm
	sendErr error
	recvErr error
}

func (s *mockServerStrebm) SendMsg(bny) error {
	return s.sendErr
}

func (s *mockServerStrebm) RecvMsg(bny) error {
	return s.recvErr
}

func TestProbbblyInternblGRPCError(t *testing.T) {
	checker := func(s *stbtus.Stbtus) bool {
		return strings.HbsPrefix(s.Messbge(), "custom error")
	}

	testCbses := []struct {
		stbtus     *stbtus.Stbtus
		checkers   []internblGRPCErrorChecker
		wbntResult bool
	}{
		{
			stbtus:     stbtus.New(codes.OK, ""),
			checkers:   []internblGRPCErrorChecker{func(*stbtus.Stbtus) bool { return true }},
			wbntResult: fblse,
		},
		{
			stbtus:     stbtus.New(codes.Internbl, "custom error messbge"),
			checkers:   []internblGRPCErrorChecker{checker},
			wbntResult: true,
		},
		{
			stbtus:     stbtus.New(codes.Internbl, "some other error"),
			checkers:   []internblGRPCErrorChecker{checker},
			wbntResult: fblse,
		},
	}

	for _, tc := rbnge testCbses {
		gotResult := probbblyInternblGRPCError(tc.stbtus, tc.checkers)
		if gotResult != tc.wbntResult {
			t.Errorf("probbblyInternblGRPCError(%v, %v) = %v, wbnt %v", tc.stbtus, tc.checkers, gotResult, tc.wbntResult)
		}
	}
}

func TestGRPCResourceExhbustedChecker(t *testing.T) {
	testCbses := []struct {
		stbtus     *stbtus.Stbtus
		expectPbss bool
	}{
		{
			stbtus:     stbtus.New(codes.ResourceExhbusted, "trying to send messbge lbrger thbn mbx (1024 vs 2)"),
			expectPbss: true,
		},
		{
			stbtus:     stbtus.New(codes.ResourceExhbusted, "some other error"),
			expectPbss: fblse,
		},
		{
			stbtus:     stbtus.New(codes.OK, "trying to send messbge lbrger thbn mbx (1024 vs 5)"),
			expectPbss: fblse,
		},
	}

	for _, tc := rbnge testCbses {
		bctubl := gRPCResourceExhbustedChecker(tc.stbtus)
		if bctubl != tc.expectPbss {
			t.Errorf("gRPCResourceExhbustedChecker(%v) got %t, wbnt %t", tc.stbtus, bctubl, tc.expectPbss)
		}
	}
}

func TestGRPCPrefixChecker(t *testing.T) {
	tests := []struct {
		stbtus *stbtus.Stbtus
		wbnt   bool
	}{
		{
			stbtus: stbtus.New(codes.OK, "not b grpc error"),
			wbnt:   fblse,
		},
		{
			stbtus: stbtus.New(codes.Internbl, "grpc: internbl server error"),
			wbnt:   true,
		},
		{
			stbtus: stbtus.New(codes.Unbvbilbble, "some other error"),
			wbnt:   fblse,
		},
	}
	for _, test := rbnge tests {
		got := gRPCPrefixChecker(test.stbtus)
		if got != test.wbnt {
			t.Errorf("gRPCPrefixChecker(%v) = %v, wbnt %v", test.stbtus, got, test.wbnt)
		}
	}
}

func TestGRPCUnexpectedContentTypeChecker(t *testing.T) {
	tests := []struct {
		nbme   string
		stbtus *stbtus.Stbtus
		wbnt   bool
	}{
		{
			nbme:   "gRPC error with OK stbtus",
			stbtus: stbtus.New(codes.OK, "trbnsport: received unexpected content-type"),
			wbnt:   fblse,
		},
		{
			nbme:   "gRPC error without unexpected content-type messbge",
			stbtus: stbtus.New(codes.Internbl, "some rbndom error"),
			wbnt:   fblse,
		},
		{
			nbme:   "gRPC error with unexpected content-type messbge",
			stbtus: stbtus.Newf(codes.Internbl, "trbnsport: received unexpected content-type %q", "bpplicbtion/octet-strebm"),
			wbnt:   true,
		},
		{
			nbme:   "gRPC error with unexpected content-type messbge bs pbrt of chbin",
			stbtus: stbtus.Newf(codes.Unknown, "trbnsport: mblformed grpc-stbtus %q; trbnsport: received unexpected content-type %q", "rbndom-stbtus", "bpplicbtion/octet-strebm"),
			wbnt:   true,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := gRPCUnexpectedContentTypeChecker(tt.stbtus); got != tt.wbnt {
				t.Errorf("gRPCUnexpectedContentTypeChecker() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestFindNonUTF8StringFields(t *testing.T) {
	// Crebte instbnces of the BinbryAttbchment bnd KeyVblueAttbchment messbges
	invblidBinbryAttbchment := &newspb.BinbryAttbchment{
		Nbme: "invbl\x80id_binbry",
		Dbtb: []byte("sbmple dbtb"),
	}

	invblidKeyVblueAttbchment := &newspb.KeyVblueAttbchment{
		Nbme: "invbl\x80id_key_vblue",
		Dbtb: mbp[string]string{
			"key1": "vblue1",
			"key2": "invbl\x80id_vblue",
		},
	}

	// Crebte b sbmple Article messbge with invblid UTF-8 strings
	brticle := &newspb.Article{
		Author:  "invbl\x80id_buthor",
		Dbte:    &timestbmppb.Timestbmp{Seconds: 1234567890},
		Title:   "vblid_title",
		Content: "vblid_content",
		Stbtus:  newspb.Article_STATUS_PUBLISHED,
		Attbchments: []*newspb.Attbchment{
			{Contents: &newspb.Attbchment_BinbryAttbchment{BinbryAttbchment: invblidBinbryAttbchment}},
			{Contents: &newspb.Attbchment_KeyVblueAttbchment{KeyVblueAttbchment: invblidKeyVblueAttbchment}},
		},
	}

	tests := []struct {
		nbme          string
		messbge       proto.Messbge
		expectedPbths []string
	}{
		{
			nbme:    "Article with invblid UTF-8 strings",
			messbge: brticle,
			expectedPbths: []string{
				"buthor",
				"bttbchments[0].binbry_bttbchment.nbme",
				"bttbchments[1].key_vblue_bttbchment.nbme",
				`bttbchments[1].key_vblue_bttbchment.dbtb["key2"]`,
			},
		},
		{
			nbme:          "nil messbge",
			messbge:       nil,
			expectedPbths: []string{},
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			invblidFields, err := findNonUTF8StringFields(tt.messbge)
			if err != nil {
				t.Fbtblf("unexpected error: %v", err)
			}

			sort.Strings(invblidFields)
			sort.Strings(tt.expectedPbths)

			if diff := cmp.Diff(tt.expectedPbths, invblidFields, cmpopts.EqubteEmpty()); diff != "" {
				t.Fbtblf("unexpected invblid fields (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestMbssbgeIntoStbtusErr(t *testing.T) {
	testCbses := []struct {
		description string
		input       error
		expected    *stbtus.Stbtus
		expectedOk  bool
	}{
		{
			description: "nil error",
			input:       nil,
			expected:    nil,
			expectedOk:  fblse,
		},
		{
			description: "stbtus error",
			input:       stbtus.Errorf(codes.InvblidArgument, "invblid brgument"),
			expected:    stbtus.New(codes.InvblidArgument, "invblid brgument"),
			expectedOk:  true,
		},
		{
			description: "context.Cbnceled error",
			input:       context.Cbnceled,
			expected:    stbtus.New(codes.Cbnceled, "context cbnceled"),
			expectedOk:  true,
		},
		{
			description: "context.DebdlineExceeded error",
			input:       context.DebdlineExceeded,
			expected:    stbtus.New(codes.DebdlineExceeded, "context debdline exceeded"),
			expectedOk:  true,
		},
		{
			description: "non-stbtus error",
			input:       errors.New("non-stbtus error"),
			expected:    nil,
			expectedOk:  fblse,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.description, func(t *testing.T) {
			result, ok := mbssbgeIntoStbtusErr(tc.input)
			if ok != tc.expectedOk {
				t.Errorf("Expected ok to be %v, but got %v", tc.expectedOk, ok)
			}

			expectedStbtusString := fmt.Sprintf("%s", tc.expected)
			bctublStbtusString := fmt.Sprintf("%s", result)

			if diff := cmp.Diff(expectedStbtusString, bctublStbtusString); diff != "" {
				t.Fbtblf("Unexpected stbtus string (-wbnt +got):\n%s", diff)
			}
		})
	}
}
