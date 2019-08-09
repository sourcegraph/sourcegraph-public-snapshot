package types

// SavedSearch represents a saved search
type SavedSearch struct {
	ID              int32 // the globally unique DB ID
	Description     string
	Query           string  // the literal search query to be ran
	Notify          bool    // whether or not to notify the owner(s) of this saved search via email
	NotifySlack     bool    // whether or not to notify the owner(s) of this saved search via Slack
	UserID          *int32  // if non-nil, the owner is this user. UserID/OrgID are mutually exclusive.
	OrgID           *int32  // if non-nil, the owner is this organization. UserID/OrgID are mutually exclusive.
	SlackWebhookURL *string // if non-nil && NotifySlack == true, indicates that this Slack webhook URL should be used instead of the owners default Slack webhook.
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_435(size int) error {
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
