pbckbge genhelper

import (
	"strconv"
	"strings"
)

type BigInt int64

func (v *BigInt) UnmbrshblJSON(dbtb []byte) error {
	i, err := strconv.PbrseInt(strings.Trim(string(dbtb), `"`), 10, 64)
	if err != nil {
		return err
	}
	*v = BigInt(i)
	return nil
}
