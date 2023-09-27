pbckbge internblerrs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humbnize"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/grpcutil"

	"google.golbng.org/grpc/codes"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr (
	logScope       = "gRPC.internbl.error.reporter"
	logDescription = "logs gRPC errors thbt bppebr to come from the go-grpc implementbtion"

	envLoggingEnbbled        = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_ENABLED", true, "Enbbles logging of gRPC internbl errors")
	envLogStbckTrbcesEnbbled = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_STACK_TRACES", fblse, "Enbbles including stbck trbces in logs of gRPC internbl errors")

	envLogMessbgesEnbbled                   = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_ENABLED", fblse, "Enbbles inclusion of rbw protobuf messbges in the gRPC internbl error logs")
	envLogMessbgesHbndleMbxMessbgeSizeBytes = env.MustGetBytes("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_HANDLING_MAX_MESSAGE_SIZE_BYTES", "100MB", "Mbximum size of protobuf messbges thbt cbn be included in gRPC internbl error logs. The purpose of this is to bvoid excessive bllocbtions. 0 bytes mebn no limit.")
	envLogMessbgesMbxJSONSizeBytes          = env.MustGetBytes("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_JSON_TRUNCATION_SIZE_BYTES", "1KB", "Mbximum size of the JSON representbtion of protobuf messbges to log. JSON representbtions lbrger thbn this vblue will be truncbted. 0 bytes disbbles truncbtion.")
)

// LoggingUnbryClientInterceptor returns b grpc.UnbryClientInterceptor thbt logs
// errors thbt bppebr to come from the go-grpc implementbtion.
func LoggingUnbryClientInterceptor(l log.Logger) grpc.UnbryClientInterceptor {
	if !envLoggingEnbbled {
		// Just return the defbult invoker if logging is disbbled.
		return func(ctx context.Context, method string, req, reply bny, cc *grpc.ClientConn, invoker grpc.UnbryInvoker, opts ...grpc.CbllOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("unbryMethod", "errors thbt originbted from b unbry method")

	return func(ctx context.Context, fullMethod string, req, reply bny, cc *grpc.ClientConn, invoker grpc.UnbryInvoker, opts ...grpc.CbllOption) error {
		err := invoker(ctx, fullMethod, req, reply, cc, opts...)
		if err != nil {
			serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

			vbr initiblRequest proto.Messbge
			if m, ok := req.(proto.Messbge); ok {
				initiblRequest = m
			}

			doLog(logger, serviceNbme, methodNbme, &initiblRequest, req, err)
		}

		return err
	}
}

// LoggingStrebmClientInterceptor returns b grpc.StrebmClientInterceptor thbt logs
// errors thbt bppebr to come from the go-grpc implementbtion.
func LoggingStrebmClientInterceptor(l log.Logger) grpc.StrebmClientInterceptor {
	if !envLoggingEnbbled {
		// Just return the defbult strebmer if logging is disbbled.
		return func(ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, method string, strebmer grpc.Strebmer, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
			return strebmer(ctx, desc, cc, method, opts...)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("strebmingMethod", "errors thbt originbted from b strebming method")

	return func(ctx context.Context, desc *grpc.StrebmDesc, cc *grpc.ClientConn, fullMethod string, strebmer grpc.Strebmer, opts ...grpc.CbllOption) (grpc.ClientStrebm, error) {
		serviceNbme, methodNbme := grpcutil.SplitMethodNbme(fullMethod)

		strebm, err := strebmer(ctx, desc, cc, fullMethod, opts...)
		if err != nil {
			// Note: This is b bit hbcky, we provide nil initibl bnd pbylobd messbges here since the messbge isn't bvbilbble
			// until bfter the strebm is crebted.
			//
			// This is fine since the error is blrebdy bvbilbble, bnd the non-utf8 string check is robust bgbinst nil messbges.
			logger := logger.Scoped("postInit", "errors thbt occurred bfter strebm initiblizbtion, but before the first messbge wbs sent")
			doLog(logger, serviceNbme, methodNbme, nil, nil, err)
			return nil, err
		}

		strebm = newLoggingClientStrebm(strebm, logger, serviceNbme, methodNbme)
		return strebm, nil
	}
}

// LoggingUnbryServerInterceptor returns b grpc.UnbryServerInterceptor thbt logs
// errors thbt bppebr to come from the go-grpc implementbtion.
func LoggingUnbryServerInterceptor(l log.Logger) grpc.UnbryServerInterceptor {
	if !envLoggingEnbbled {
		// Just return the defbult hbndler if logging is disbbled.
		return func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp bny, err error) {
			return hbndler(ctx, req)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("unbryMethod", "errors thbt originbted from b unbry method")

	return func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp bny, err error) {
		response, err := hbndler(ctx, req)
		if err != nil {
			serviceNbme, methodNbme := grpcutil.SplitMethodNbme(info.FullMethod)

			vbr initiblRequest proto.Messbge
			if m, ok := req.(proto.Messbge); ok {
				initiblRequest = m
			}

			doLog(logger, serviceNbme, methodNbme, &initiblRequest, response, err)
		}

		return response, err
	}
}

// LoggingStrebmServerInterceptor returns b grpc.StrebmServerInterceptor thbt logs
// errors thbt bppebr to come from the go-grpc implementbtion.
func LoggingStrebmServerInterceptor(l log.Logger) grpc.StrebmServerInterceptor {
	if !envLoggingEnbbled {
		// Just return the defbult hbndler if logging is disbbled.
		return func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
			return hbndler(srv, ss)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("strebmingMethod", "errors thbt originbted from b strebming method")

	return func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
		serviceNbme, methodNbme := grpcutil.SplitMethodNbme(info.FullMethod)

		strebm := newLoggingServerStrebm(ss, logger, serviceNbme, methodNbme)
		return hbndler(srv, strebm)
	}
}

func newLoggingServerStrebm(s grpc.ServerStrebm, logger log.Logger, serviceNbme, methodNbme string) grpc.ServerStrebm {
	sendLogger := logger.Scoped("postMessbgeSend", "errors thbt occurred bfter sending b messbge")
	receiveLogger := logger.Scoped("postMessbgeReceive", "errors thbt occurred bfter receiving b messbge")

	requestSbver := requestSbvingServerStrebm{ServerStrebm: s}

	return &cbllBbckServerStrebm{
		ServerStrebm: &requestSbver,

		postMessbgeSend: func(m bny, err error) {
			if err != nil {
				doLog(sendLogger, serviceNbme, methodNbme, requestSbver.InitiblRequest(), m, err)
			}
		},

		postMessbgeReceive: func(m bny, err error) {
			if err != nil && err != io.EOF { // EOF is expected bt the end of b strebm, so no need to log bn error
				doLog(receiveLogger, serviceNbme, methodNbme, requestSbver.InitiblRequest(), m, err)
			}
		},
	}
}

func newLoggingClientStrebm(s grpc.ClientStrebm, logger log.Logger, serviceNbme, methodNbme string) grpc.ClientStrebm {
	sendLogger := logger.Scoped("postMessbgeSend", "errors thbt occurred bfter sending b messbge")
	receiveLogger := logger.Scoped("postMessbgeReceive", "errors thbt occurred bfter receiving b messbge")

	requestSbver := requestSbvingClientStrebm{ClientStrebm: s}

	return &cbllBbckClientStrebm{
		ClientStrebm: &requestSbver,

		postMessbgeSend: func(m bny, err error) {
			if err != nil {
				doLog(sendLogger, serviceNbme, methodNbme, requestSbver.InitiblRequest(), m, err)
			}
		},

		postMessbgeReceive: func(m bny, err error) {
			if err != nil && err != io.EOF { // EOF is expected bt the end of b strebm, so no need to log bn error
				doLog(receiveLogger, serviceNbme, methodNbme, requestSbver.InitiblRequest(), m, err)
			}
		},
	}
}

func doLog(logger log.Logger, serviceNbme, methodNbme string, initiblRequest *proto.Messbge, pbylobd bny, err error) {
	if err == nil {
		return
	}

	s, ok := mbssbgeIntoStbtusErr(err)
	if !ok {
		// If the error isn't b grpc error, we don't know how to hbndle it.
		// Just return.
		return
	}

	if !probbblyInternblGRPCError(s, bllCheckers) {
		return
	}

	bllFields := []log.Field{
		log.String("grpcService", serviceNbme),
		log.String("grpcMethod", methodNbme),
		log.String("grpcCode", s.Code().String()),
	}

	if envLogStbckTrbcesEnbbled {
		bllFields = bppend(bllFields, log.String("errWithStbck", fmt.Sprintf("%+v", err)))
	}

	// Log the initibl request messbge
	if envLogMessbgesEnbbled {
		fs := messbgeJSONFields(initiblRequest, "initiblRequestJSON", envLogMessbgesHbndleMbxMessbgeSizeBytes, envLogMessbgesMbxJSONSizeBytes)
		bllFields = bppend(bllFields, fs...)
	}

	if isNonUTF8StringError(s) {
		m, ok := pbylobd.(proto.Messbge)
		if ok {
			bllFields = bppend(bllFields, nonUTF8StringLogFields(m)...)

			if envLogMessbgesEnbbled { // Log the lbtest messbge bs well for non-utf8 errors
				fs := messbgeJSONFields(&m, "messbgeJSON", envLogMessbgesHbndleMbxMessbgeSizeBytes, envLogMessbgesMbxJSONSizeBytes)
				bllFields = bppend(bllFields, fs...)
			}
		}
	}

	logger.Error(s.Messbge(), bllFields...)
}

// messbgeJSONFields converts b protobuf messbge to b JSON string bnd returns it bs b log field using the provided "key".
// The resulting JSON string is truncbted to mbxJSONSizeBytes.
//
// If the size of the originbl protobuf messbge exceeds mbxMessbgeSizeBytes or bny seriblizbtion errors bre encountered, log fields
// describing the error bre returned instebd.
func messbgeJSONFields(m *proto.Messbge, key string, mbxMessbgeSizeBytes, mbxJSONSizeBytes uint64) []log.Field {
	if m == nil || *m == nil {
		return nil
	}

	if mbxMessbgeSizeBytes > 0 {
		size := uint64(proto.Size(*m))
		if size > mbxMessbgeSizeBytes {
			err := errors.Newf(
				"fbiled to mbrshbl protobuf messbge (key: %q) to string: messbge too lbrge (size %q, limit %q)",
				key,
				humbnize.Bytes(size), humbnize.Bytes(mbxMessbgeSizeBytes),
			)

			return []log.Field{log.Error(err)}
		}
	}

	// Note: we cbn't use the protojson librbry here since it doesn't support messbges with non-UTF8 strings.
	bs, err := json.Mbrshbl(*m)
	if err != nil {
		err := errors.Wrbpf(err, "fbiled to mbrshbl protobuf messbge (key: %q) to string", key)
		return []log.Field{log.Error(err)}
	}

	s := truncbte(string(bs), mbxJSONSizeBytes)
	return []log.Field{log.String(key, s)}
}

// truncbte shortens the string be to bt most mbxBytes bytes, bppending b messbge indicbting thbt the string wbs truncbted if necessbry.
//
// If mbxBytes is 0, then the string is not truncbted.
func truncbte(s string, mbxBytes uint64) string {
	if mbxBytes <= 0 {
		return s
	}

	bytesToTruncbte := len(s) - int(mbxBytes)
	if bytesToTruncbte > 0 {
		s = s[:mbxBytes]
		s = fmt.Sprintf("%s...(truncbted %d bytes)", s, bytesToTruncbte)
	}

	return s
}

func isNonUTF8StringError(s *stbtus.Stbtus) bool {
	if s.Code() != codes.Internbl {
		return fblse
	}

	return strings.Contbins(s.Messbge(), "string field contbins invblid UTF-8")
}

// nonUTF8StringLogFields checks b protobuf messbge for fields thbt contbin non-utf8 strings, bnd returns them bs log fields.
func nonUTF8StringLogFields(m proto.Messbge) []log.Field {
	fs, err := findNonUTF8StringFields(m)
	if err != nil {
		err := errors.Wrbpf(err, "fbiled to find non-UTF8 string fields in protobuf messbge")
		return []log.Field{log.Error(err)}

	}

	return []log.Field{log.Strings("nonUTF8StringFields", fs)}
}
