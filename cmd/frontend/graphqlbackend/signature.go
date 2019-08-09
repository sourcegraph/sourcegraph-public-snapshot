package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

type signatureResolver struct {
	person *personResolver
	date   time.Time
}

func (r signatureResolver) Person() *personResolver {
	return r.person
}

func (r signatureResolver) Date() string {
	return r.date.Format(time.RFC3339)
}

func toSignatureResolver(sig *git.Signature) *signatureResolver {
	if sig == nil {
		return nil
	}
	return &signatureResolver{
		person: &personResolver{
			name:  sig.Name,
			email: sig.Email,
		},
		date: sig.Date,
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_216(size int) error {
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
