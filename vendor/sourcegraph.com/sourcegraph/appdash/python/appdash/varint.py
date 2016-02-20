# Hack: There is no official EncodeVarint function provided by protobuf. The
# one we use here is technically private. If it is removed or becomes
# troublesome to update -- we'll have to do it ourselves.
#
# See http://code.google.com/p/protobuf/issues/detail?id=226
import google.protobuf.internal.encoder as enc
from google.protobuf.internal.encoder import _EncodeVarint

# encode encodes the unsigned varint i, and returns the encoded data as
# a str.
def encode(i):
    buf = []
	# Note: The signed variant is named "_EncodeSignedVarint".
    _EncodeVarint(buf.append, i)
    return "".join(buf)

