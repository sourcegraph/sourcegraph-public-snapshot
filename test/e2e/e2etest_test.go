package e2e

import (
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestRepoFlow(t *testing.T) {
	runE2E(t, "repo_flow")
}

func TestSearchFlow0(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1292")
	runE2E(t, "search_flow_0")
}

func TestSearchFlow1(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1292")
	runE2E(t, "search_flow_1")
}

func TestSearchFlow2(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1292")
	runE2E(t, "search_flow_2")
}

func TestSearchFlow3(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1292")
	runE2E(t, "search_flow_3")
}

func TestSearchFlow4(t *testing.T) {
	t.Skip("Disabled https://github.com/sourcegraph/sourcegraph/issues/1292")
	runE2E(t, "search_flow_4")
}

var fatalMsg string

func TestMain(m *testing.M) {
	flag.Parse()

	// Prepare logging.
	tr.log = log.New(io.MultiWriter(os.Stderr, tr.slackLogBuffer), "", 0)

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
	usernamePrefix = usernamePrefix + hwid

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

	err, screenshot := tr.runTest(test, t)
	if err != nil {
		if screenshots := os.Getenv("WRITE_SCREENSHOTS"); screenshots != "" {
			err2 := os.MkdirAll(screenshots, 0700)
			if err2 != nil {
				t.Fatal(err2)
			}
			path, err2 := filepath.Abs(filepath.Join(screenshots, test.Name+".png"))
			if err2 != nil {
				t.Fatal(err2)
			}
			if screenshot == nil {
				t.Logf("Skipped writing empty screenshot to %s\n", path)
			} else {
				err2 = ioutil.WriteFile(path, screenshot, 0666)
				if err2 != nil {
					t.Fatal(err2)
				}
				t.Logf("Wrote: %s\n", path)
			}
		}
		t.Fatal(err)
	}
}
