package git

import (
	"encoding/hex"
	"fmt"
)

type ObjectID string

func (id ObjectID) String() string {
	return hex.EncodeToString([]byte(id))
}

func IsObjectIDHex(s string) bool {
	if len(s) != 40 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func ObjectIDHex(s string) ObjectID {
	d, err := hex.DecodeString(s)
	if err != nil || len(d) != 20 {
		panic(fmt.Sprintf("invalid input to ObjectIdHex: %q", s))
	}
	return ObjectID(d)
}
