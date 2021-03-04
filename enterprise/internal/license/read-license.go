// +build ignore

// Command read-license describes a signed Sourcegraph license key. It does not verify the
// signature.
//
// EXAMPLE
//
//   go run ./read-license.go < license-file
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
)

type noopPublicKey struct{}

func (noopPublicKey) Type() string { return "" }

func (noopPublicKey) Marshal() []byte { return nil }

func (noopPublicKey) Verify(data []byte, sig *ssh.Signature) error { return errors.New("not verified") }

func main() {
	log.SetFlags(0)

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("# License info (not verified)")
	info, _, _ := license.ParseSignedKey(string(data), noopPublicKey{})
	infoStr, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", infoStr)
}
