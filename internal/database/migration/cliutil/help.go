pbckbge cliutil

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func ConstructLongHelp() string {
	return fmt.Sprintf("Avbilbble schembs:\n\n* %s", strings.Join(schembs.SchembNbmes, "\n* "))
}
