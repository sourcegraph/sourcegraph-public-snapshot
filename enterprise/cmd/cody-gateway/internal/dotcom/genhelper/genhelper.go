package genhelper

import (
	"strconv"
	"strings"
)

type BigInt int64

func (v *BigInt) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(strings.Trim(string(data), `"`), 10, 64)
	if err != nil {
		return err
	}
	*v = BigInt(i)
	return nil
}
