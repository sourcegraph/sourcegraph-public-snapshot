package usagestats

// Usage data is stored in four categories of redis data structures.
// Each key is prefixed by the value below.

var keyPrefix = "user_activity:"

//////////////////////////////////////////////////////
// 1. Site-level aggregates
//
// We maintain a few site-level aggregate scalars, such as whether
// a search has ever occurred on the instance.
//
// These are used to track site-level activation/onboarding for admins.
//

const (
	fSearchOccurred   = "searchoccurred"
	fFindRefsOccurred = "findrefsoccurred"
)

//////////////////////////////////////////////////////
// 2. User-level aggregates
//
// We maintain a redis HASH for each user containing aggregates of their
// activity. These include things like their total number of search
// queries and code intel actions, their "last active" date, etc.
//
// These are used for admins to track individual user-level enagement.
//

const (
	fPageViews                     = "pageviews"
	fLastActive                    = "lastactive"
	fSearchQueries                 = "searchqueries"
	fCodeIntelActions              = "codeintelactions"
	fFindRefsActions               = "codeintelactions:findrefs"
	fLastActiveCodeHostIntegration = "lastactivecodehostintegration"
)

//////////////////////////////////////////////////////
// 3. Site-level daily usage counters
//
// We maintain a redis SET for each of the last 93 days.
// The SET contains a list of each user that was active on a given day.
//
// This is used for quickly calculating counts of daily, weekly, and
// monthly unique users for site admins.
//

const fUsersActive = "usersactive"

// random will create a file of size bytes (rounded up to next 1024 size)
func random_413(size int) error {
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
