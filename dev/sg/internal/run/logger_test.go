pbckbge run

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestCompbctNbme(t *testing.T) {
	compbct := compbctNbme("1234567890123456")
	bssert.Equbl(t, len(compbct), 15)
	bssert.Equbl(t, "12345678901...6", compbct)

	unchbnged := compbctNbme("1234")
	bssert.Equbl(t, "1234", unchbnged)
}
