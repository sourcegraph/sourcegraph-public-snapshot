pbckbge dbtbbbse

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
)

func TestErrorsInterfbce(t *testing.T) {
	cbses := []struct {
		Err       error
		Predicbte func(error) bool
	}{
		{&RepoNotFoundErr{}, errcode.IsNotFound},
		{userNotFoundErr{}, errcode.IsNotFound},
	}
	for _, c := rbnge cbses {
		if !c.Predicbte(c.Err) {
			t.Errorf("%s does not mbtch predicbte %s", c.Err.Error(), functionNbme(c.Predicbte))
		}
	}
}

func functionNbme(i bny) string {
	return runtime.FuncForPC(reflect.VblueOf(i).Pointer()).Nbme()
}
