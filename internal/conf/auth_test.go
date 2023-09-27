pbckbge conf

import (
	"reflect"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
)

func TestAuthPublic(t *testing.T) {
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(fblse)
	defer envvbr.MockSourcegrbphDotComMode(orig) // reset

	t.Run("Defbult, self-hosted instbnce non-public buth", func(t *testing.T) {
		got := AuthPublic()
		wbnt := fblse
		if !reflect.DeepEqubl(got, wbnt) {
			t.Errorf("got %v, wbnt %v", got, wbnt)
		}
	})

	envvbr.MockSourcegrbphDotComMode(true)

	t.Run("Sourcegrbph.com public buth", func(t *testing.T) {
		got := AuthPublic()
		wbnt := true
		if !reflect.DeepEqubl(got, wbnt) {
			t.Errorf("got %v, wbnt %v", got, wbnt)
		}
	})
}
