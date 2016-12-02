package store

import "encoding/json"

func deepEqual(u, v interface{}) bool {
	u_, _ := json.Marshal(u)
	v_, _ := json.Marshal(v)
	return string(u_) == string(v_)
}
