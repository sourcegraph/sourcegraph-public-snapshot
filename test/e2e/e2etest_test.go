package e2e

import (
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"os"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/test/e2e/e2etestuser"
)

func TestDefFlow(t *testing.T) {
	runE2E(t, "def_flow")
}

func TestLoginFlow(t *testing.T) {
	runE2E(t, "login_flow")
}

func TestRegisterFlow(t *testing.T) {
	runE2E(t, "register_flow")
}

func TestRepoFlow(t *testing.T) {
	runE2E(t, "repo_flow")
}

func TestSearchFlow(t *testing.T) {
	runE2E(t, "search_flow")
}

func TestChannelFlow(t *testing.T) {
	runE2E(t, "channel_flow")
}

var fatalMsg string

func TestMain(m *testing.M) {
	flag.Parse()
	err := parseEnv()
	if err != nil {
		fatalMsg = "parseEnv: " + err.Error()
	}

	// Prevent collision between multiple people running e2etest at the same
	// time against the same target instance. Otherwise, we may hit race
	// conditions / etc.
	hwid, err := HardwareID()
	if err != nil {
		log.Fatal(err)
	}
	e2etestuser.Prefix = e2etestuser.Prefix + hwid

	seleniumTrace = testing.Verbose()
	os.Exit(m.Run())
}

// HardwareID returns a CRC32 checksum of the first network adapter with an
// hardware address.
func HardwareID() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if addr := iface.HardwareAddr.String(); addr != "" {
			return fmt.Sprint(crc32.ChecksumIEEE([]byte(addr))), nil
		}
	}
	return "", errors.New("could not find a hardware address")
}

func runE2E(t *testing.T, name string) {
	t.Parallel()
	var test *Test
	for _, tst := range tr.tests {
		if tst.Name == name {
			test = tst
		}
	}
	if test == nil {
		t.Fatal("Could not find test")
	}
	if fatalMsg != "" {
		t.Fatal(fatalMsg)
	}
	wd, err := tr.newWebDriver()
	if err != nil {
		t.Skip("newWebDriver:", err)
	}
	defer wd.Quit()
	e2eT := tr.newT(test, wd)
	e2eT.testingT = t
	err = test.Func(e2eT)
	if err != nil {
		t.Fatal(err)
	}
}
