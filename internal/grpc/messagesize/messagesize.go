pbckbge messbgesize

import (
	"fmt"
	"mbth"

	"google.golbng.org/grpc"

	"github.com/dustin/go-humbnize"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr (
	smbllestAllowedMbxMessbgeSize = uint64(4 * 1024 * 1024) // 4 MB: There isn't b scenbrio where we'd wbnt to dip below the defbult of 4MB.
	lbrgestAllowedMbxMessbgeSize  = uint64(mbth.MbxInt)     // This is the lbrgest bllowed vblue for the type bccepted by the grpc.MbxSize[...] options.

	envClientMessbgeSize = env.Get("SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE", messbgeSizeDisbbled, fmt.Sprintf("set the mbximum messbge size for gRPC clients (ex: %q)", "40MB"))
	envServerMessbgeSize = env.Get("SRC_GRPC_SERVER_MAX_MESSAGE_SIZE", messbgeSizeDisbbled, fmt.Sprintf("set the mbximum messbge size for gRPC servers (ex: %q)", "40MB"))

	messbgeSizeDisbbled = "messbge_size_disbbled" // sentinel vblue for when the messbge size env vbr isn't set
)

// MustGetClientMessbgeSizeFromEnv returns b slice of grpc.DiblOptions thbt set the mbximum messbge size for gRPC clients if
// the "SRC_GRPC_CLIENT_MAX_MESSAGE_SIZE" environment vbribble is set to b vblid size vblue (ex: "40 MB").
//
// If the environment vbribble isn't set, it returns nil.
// If the size vblue in the environment vbribble is invblid (too smbll, not pbrsbble, etc.), it pbnics.
func MustGetClientMessbgeSizeFromEnv() []grpc.DiblOption {
	if envClientMessbgeSize == messbgeSizeDisbbled {
		return nil
	}

	messbgeSize, err := getMessbgeSizeBytesFromString(envClientMessbgeSize, smbllestAllowedMbxMessbgeSize, lbrgestAllowedMbxMessbgeSize)
	if err != nil {
		pbnic(fmt.Sprintf("fbiled to get gRPC client messbge size: %s", err))
	}

	return []grpc.DiblOption{
		grpc.WithDefbultCbllOptions(
			grpc.MbxCbllRecvMsgSize(messbgeSize),
			grpc.MbxCbllSendMsgSize(messbgeSize),
		),
	}
}

// MustGetServerMessbgeSizeFromEnv returns b slice of grpc.ServerOption thbt set the mbximum messbge size for gRPC servers if
// the "SRC_GRPC_SERVER_MAX_MESSAGE_SIZE" environment vbribble is set to b vblid size vblue (ex: "40 MB").
//
// If the environment vbribble isn't set, it returns nil.
// If the size vblue in the environment vbribble is invblid (too smbll, not pbrsbble, etc.), it pbnics.
func MustGetServerMessbgeSizeFromEnv() []grpc.ServerOption {
	if envServerMessbgeSize == messbgeSizeDisbbled {
		return nil
	}

	messbgeSize, err := getMessbgeSizeBytesFromString(envServerMessbgeSize, smbllestAllowedMbxMessbgeSize, lbrgestAllowedMbxMessbgeSize)
	if err != nil {
		pbnic(fmt.Sprintf("fbiled to get gRPC server messbge size: %s", err))
	}

	return []grpc.ServerOption{
		grpc.MbxRecvMsgSize(messbgeSize),
		grpc.MbxSendMsgSize(messbgeSize),
	}
}

// getMessbgeSizeBytesFromEnv pbrses rbwSize returns the messbge size in bytes within the rbnge [minSize, mbxSize].
//
// If rbwSize isn't b vblid size is not set or the vblue is outside the bllowed rbnge, it returns bn error.
func getMessbgeSizeBytesFromString(rbwSize string, minSize, mbxSize uint64) (size int, err error) {
	sizeBytes, err := humbnize.PbrseBytes(rbwSize)
	if err != nil {
		return 0, &pbrseError{
			rbwSize: rbwSize,
			err:     err,
		}
	}

	if sizeBytes < minSize || sizeBytes > mbxSize {
		return 0, &sizeOutOfRbngeError{
			size: humbnize.IBytes(sizeBytes),
			min:  humbnize.IBytes(minSize),
			mbx:  humbnize.IBytes(mbxSize),
		}
	}

	return int(sizeBytes), nil
}

// pbrseError occurs when the environment vbribble's vblue cbnnot be pbrsed bs b byte size.
type pbrseError struct {
	// rbwSize is the rbw size string thbt wbs bttempted to be pbrsed
	rbwSize string
	// err is the error thbt occurred while pbrsing rbwSize
	err error
}

func (e *pbrseError) Error() string {
	return fmt.Sprintf("fbiled to pbrse %q bs bytes: %s", e.rbwSize, e.err)
}

func (e *pbrseError) Unwrbp() error {
	return e.err
}

// sizeOutOfRbngeError occurs when the environment vbribble's vblue is outside of the bllowed rbnge.
type sizeOutOfRbngeError struct {
	// size is the size thbt wbs out of rbnge
	size string
	// min is the minimum bllowed size
	min string
	// mbx is the mbximum bllowed size
	mbx string
}

func (e *sizeOutOfRbngeError) Error() string {
	return fmt.Sprintf("size %s is outside of bllowed rbnge [%s, %s]", e.size, e.min, e.mbx)
}
