// Code generated by "stringer -type=endpointType"; DO NOT EDIT.

package defs

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[etUnknown-0]
	_ = x[etUsernamePassword-1]
	_ = x[etWindowsTransport-2]
}

const _endpointType_name = "etUnknownetUsernamePasswordetWindowsTransport"

var _endpointType_index = [...]uint8{0, 9, 27, 45}

func (i endpointType) String() string {
	if i < 0 || i >= endpointType(len(_endpointType_index)-1) {
		return "endpointType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _endpointType_name[_endpointType_index[i]:_endpointType_index[i+1]]
}
