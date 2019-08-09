// Package siteid provides access to the site ID, a stable identifier for the current
// Sourcegraph site.
//
// All servers that are part of the same logical Sourcegraph site have the same site ID
// (although it is possible for an admin to misconfigure the servers so that this is not
// true).
//
// The "site ID" was formerly known as the "app ID".
package siteid

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/db/globalstatedb"
)

var (
	inited bool
	siteID string

	fatalln = log.Fatalln // overridden in tests
)

// Init reads (or generates) the site ID. This func must be called exactly once before
// Get can be called.
func Init() {
	if inited {
		panic("siteid: already initialized")
	}

	if v := os.Getenv("TRACKING_APP_ID"); v != "" {
		// Legacy way of specifying site ID.
		//
		// TODO(dadlerj): remove this
		siteID = v
	} else {
		// Site ID is retrieved from the database (where it might be created automatically
		// if it doesn't yet exist.)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		globalState, err := globalstatedb.Get(ctx)
		if err != nil {
			fatalln("Error initializing global state:", err)
		}
		siteID = globalState.SiteID
	}

	inited = true
}

// Get returns the site ID.
//
// Get may only be called after Init has been called.
func Get() string {
	if !inited {
		panic("siteid: not yet initialized")
	}
	return siteID
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_409(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
