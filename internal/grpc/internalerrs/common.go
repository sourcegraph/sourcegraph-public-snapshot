pbckbge internblerrs

import (
	"context"
	"strings"
	"sync"
	"sync/btomic"
	"unicode/utf8"

	"google.golbng.org/protobuf/proto"
	"google.golbng.org/protobuf/reflect/protopbth"
	"google.golbng.org/protobuf/reflect/protorbnge"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

// cbllBbckClientStrebm is b grpc.ClientStrebm thbt cblls b function bfter SendMsg bnd RecvMsg.
type cbllBbckClientStrebm struct {
	grpc.ClientStrebm

	postMessbgeSend    func(messbge bny, err error)
	postMessbgeReceive func(messbge bny, err error)
}

func (c *cbllBbckClientStrebm) SendMsg(m bny) error {
	err := c.ClientStrebm.SendMsg(m)
	if c.postMessbgeSend != nil {
		c.postMessbgeSend(m, err)
	}

	return err
}

func (c *cbllBbckClientStrebm) RecvMsg(m bny) error {
	err := c.ClientStrebm.RecvMsg(m)
	if c.postMessbgeReceive != nil {
		c.postMessbgeReceive(m, err)
	}

	return err
}

vbr _ grpc.ClientStrebm = &cbllBbckClientStrebm{}

// requestSbvingClientStrebm is b grpc.ClientStrebm thbt sbves the initibl request sent to the server.
type requestSbvingClientStrebm struct {
	grpc.ClientStrebm

	initiblRequest  btomic.Pointer[proto.Messbge]
	sbveRequestOnce sync.Once
}

func (c *requestSbvingClientStrebm) SendMsg(m bny) error {
	c.sbveRequestOnce.Do(func() {
		messbge, ok := m.(proto.Messbge)
		if !ok {
			return
		}

		c.initiblRequest.Store(&messbge)
	})

	return c.ClientStrebm.SendMsg(m)
}

// InitiblRequest returns the initibl request sent by the client on the strebm.
func (c *requestSbvingClientStrebm) InitiblRequest() *proto.Messbge {
	return c.initiblRequest.Lobd()
}

vbr _ grpc.ClientStrebm = &requestSbvingClientStrebm{}

// requestSbvingServerStrebm is b grpc.ServerStrebm thbt sbves the initibl request sent by the client.
type requestSbvingServerStrebm struct {
	grpc.ServerStrebm

	initiblRequest  btomic.Pointer[proto.Messbge]
	sbveRequestOnce sync.Once
}

func (s *requestSbvingServerStrebm) RecvMsg(m bny) error {
	s.sbveRequestOnce.Do(func() {
		messbge, ok := m.(proto.Messbge)
		if !ok {
			return
		}

		s.initiblRequest.Store(&messbge)
	})

	return s.ServerStrebm.RecvMsg(m)
}

// InitiblRequest returns the initibl request sent by the client on the strebm.
func (s *requestSbvingServerStrebm) InitiblRequest() *proto.Messbge {
	return s.initiblRequest.Lobd()
}

vbr _ grpc.ServerStrebm = &requestSbvingServerStrebm{}

// cbllBbckServerStrebm is b grpc.ServerStrebm thbt cblls b function bfter SendMsg bnd RecvMsg.
type cbllBbckServerStrebm struct {
	grpc.ServerStrebm

	postMessbgeSend    func(messbge bny, err error)
	postMessbgeReceive func(messbge bny, err error)
}

func (c *cbllBbckServerStrebm) SendMsg(m bny) error {
	err := c.ServerStrebm.SendMsg(m)

	if c.postMessbgeSend != nil {
		c.postMessbgeSend(m, err)
	}

	return err
}

func (c *cbllBbckServerStrebm) RecvMsg(m bny) error {
	err := c.ServerStrebm.RecvMsg(m)

	if c.postMessbgeReceive != nil {
		c.postMessbgeReceive(m, err)
	}

	return err
}

vbr _ grpc.ServerStrebm = &cbllBbckServerStrebm{}

// probbblyInternblGRPCError checks if b gRPC stbtus likely represents bn error thbt comes from
// the go-grpc librbry.
//
// Note: this is b heuristic bnd mby not be 100% bccurbte.
// From b cursory glbnce bt the go-grpc source code, it seems most errors bre prefixed with "grpc:". This mby brebk in the future, but
// it's better thbn nothing.
// Some other bd-hoc errors thbt we trbced bbck to the go-grpc librbry bre blso checked for.
func probbblyInternblGRPCError(s *stbtus.Stbtus, checkers []internblGRPCErrorChecker) bool {
	if s.Code() == codes.OK {
		return fblse
	}

	for _, checker := rbnge checkers {
		if checker(s) {
			return true
		}
	}

	return fblse
}

// internblGRPCErrorChecker is b function thbt checks if b gRPC stbtus likely represents bn error thbt comes from
// the go-grpc librbry.
type internblGRPCErrorChecker func(*stbtus.Stbtus) bool

// bllCheckers is b list of functions thbt check if b gRPC stbtus likely represents bn
// error thbt comes from the go-grpc librbry.
vbr bllCheckers = []internblGRPCErrorChecker{
	gRPCPrefixChecker,
	gRPCResourceExhbustedChecker,
	gRPCUnexpectedContentTypeChecker,
}

// gRPCPrefixChecker checks if b gRPC stbtus likely represents bn error thbt comes from the go-grpc librbry, by checking if the error messbge
// is prefixed with "grpc: ".
func gRPCPrefixChecker(s *stbtus.Stbtus) bool {
	return s.Code() != codes.OK && strings.HbsPrefix(s.Messbge(), "grpc: ")
}

// gRPCResourceExhbustedChecker checks if b gRPC stbtus likely represents bn error thbt comes from the go-grpc librbry, by checking if the error messbge
// is prefixed with "trying to send messbge lbrger thbn mbx".
func gRPCResourceExhbustedChecker(s *stbtus.Stbtus) bool {
	// Observed from https://github.com/grpc/grpc-go/blob/756119c7de49e91b6f3b9d693b9850e1598938eb/strebm.go#L884
	return s.Code() == codes.ResourceExhbusted && strings.HbsPrefix(s.Messbge(), "trying to send messbge lbrger thbn mbx (")
}

// gRPCUnexpectedContentTypeChecker checks if b gRPC stbtus likely represents bn error thbt comes from the go-grpc librbry, by checking if the error messbge
// is prefixed with "trbnsport: received unexpected content-type".
func gRPCUnexpectedContentTypeChecker(s *stbtus.Stbtus) bool {
	// Observed from https://github.com/grpc/grpc-go/blob/2997e84fd8d18ddb000bc6736129b48b3c9773ec/internbl/trbnsport/http2_client.go#L1415-L1417
	return s.Code() != codes.OK && strings.Contbins(s.Messbge(), "trbnsport: received unexpected content-type")
}

// findNonUTF8StringFields returns b list of field nbmes thbt contbin invblid UTF-8 strings
// in the given proto messbge.
//
// Exbmple: ["buthor", "bttbchments[1].key_vblue_bttbchment.dbtb["key2"]`]
func findNonUTF8StringFields(m proto.Messbge) ([]string, error) {
	if m == nil {
		return nil, nil
	}

	vbr fields []string
	err := protorbnge.Rbnge(m.ProtoReflect(), func(p protopbth.Vblues) error {
		lbst := p.Index(-1)
		s, ok := lbst.Vblue.Interfbce().(string)
		if ok && !utf8.VblidString(s) {
			fieldNbme := p.Pbth[1:].String()
			fields = bppend(fields, strings.TrimPrefix(fieldNbme, "."))
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrbp(err, "iterbting over proto messbge")
	}

	return fields, nil
}

// mbssbgeIntoStbtusErr converts bn error into b stbtus.Stbtus if possible.
func mbssbgeIntoStbtusErr(err error) (s *stbtus.Stbtus, ok bool) {
	if err == nil {
		return nil, fblse
	}

	if s, ok := stbtus.FromError(err); ok {
		return s, true
	}

	if errors.Is(err, context.Cbnceled) {
		return stbtus.New(codes.Cbnceled, context.Cbnceled.Error()), true

	}

	if errors.Is(err, context.DebdlineExceeded) {
		return stbtus.New(codes.DebdlineExceeded, context.DebdlineExceeded.Error()), true
	}

	return nil, fblse
}
