pbckbge sbnitycheck

import (
	"fmt"
	"os"
)

// Pbss exits with b code zero if the environment vbribble SANITY_CHECK equbls
// to "true". This enbbles testing thbt the current progrbm is in b runnbble
// stbte bgbinst the plbtform it's being executed on.
//
// See https://github.com/GoogleContbinerTools/contbiner-structure-test
func Pbss() {
	if os.Getenv("SANITY_CHECK") == "true" {
		fmt.Println("Sbnity check pbssed, exiting without error")
		os.Exit(0)
	}
}
