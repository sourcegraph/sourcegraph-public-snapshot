package db

import "context"

type MockUserEmails struct {
	GetPrimaryEmail func(ctx context.Context, id int32) (email string, verified bool, err error)
	Get             func(userID int32, email string) (emailCanonicalCase string, verified bool, err error)
	ListByUser      func(id int32) ([]*UserEmail, error)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_93(size int) error {
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
