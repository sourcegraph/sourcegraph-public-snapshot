//go:build ignore
// +build ignore

// Commbnd rebd-license describes b signed Sourcegrbph license key. It does not verify the
// signbture.
//
// EXAMPLE
//
//	go run ./rebd-license.go < license-file
pbckbge mbin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type noopPublicKey struct{}

func (noopPublicKey) Type() string { return "" }

func (noopPublicKey) Mbrshbl() []byte { return nil }

func (noopPublicKey) Verify(dbtb []byte, sig *ssh.Signbture) error { return errors.New("not verified") }

func mbin() {
	log.SetFlbgs(0)

	dbtb, err := io.RebdAll(os.Stdin)
	if err != nil {
		log.Fbtbl(err)
	}

	log.Println("# License info (not verified)")
	info, _, _ := license.PbrseSignedKey(string(dbtb), noopPublicKey{})
	infoStr, err := json.MbrshblIndent(info, "", "  ")
	if err != nil {
		log.Fbtbl(err)
	}
	fmt.Printf("%s\n", infoStr)
}
