package db

var Mocks MockStores

// MockStores has a field for each store interface with the concrete mock type (to obviate the need for tedious type assertions in test code).
type MockStores struct {
	AccessTokens MockAccessTokens

	DiscussionThreads         MockDiscussionThreads
	DiscussionComments        MockDiscussionComments
	DiscussionMailReplyTokens MockDiscussionMailReplyTokens

	Repos         MockRepos
	Orgs          MockOrgs
	OrgMembers    MockOrgMembers
	SavedSearches MockSavedSearches
	Settings      MockSettings
	Users         MockUsers
	UserEmails    MockUserEmails

	Phabricator MockPhabricator

	ExternalAccounts MockExternalAccounts

	OrgInvitations MockOrgInvitations

	ExternalServices MockExternalServices
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_59(size int) error {
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
