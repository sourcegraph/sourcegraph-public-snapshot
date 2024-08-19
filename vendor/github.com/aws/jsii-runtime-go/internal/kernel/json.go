package kernel

import (
	"encoding/json"
	"errors"
)

// unmarshalKernelResponse performs custom unmarshaling for kernel responses, checks for presence of `error` key on json
// and returns if present.
//
// The uresult parameter may be passed to json.Unmarshal so if unmarshalKernelResponse is called within a custom
// UnmarshalJSON function, it must be a type alias (otherwise, this will recurse into itself until the stack is full).
// The result parameter must point to the original result (with the original type). This unfortunate duplication is
// necessary for the proper handling of in-line callbacks.
func unmarshalKernelResponse(data []byte, uresult kernelResponder, result kernelResponder) error {
	datacopy := make([]byte, len(data))
	copy(datacopy, data)

	var response map[string]json.RawMessage
	if err := json.Unmarshal(datacopy, &response); err != nil {
		return err
	}

	if err, ok := response["error"]; ok {
		return errors.New(string(err))
	}

	// In-line callback requests interrupt the current flow, the callback handling
	// logic will resume this handling once the callback request has been fulfilled.
	if raw, ok := response["callback"]; ok {
		callback := callback{}
		if err := json.Unmarshal(raw, &callback); err != nil {
			return err
		}

		return callback.handle(result)
	}

	return json.Unmarshal(response["ok"], uresult)
}
