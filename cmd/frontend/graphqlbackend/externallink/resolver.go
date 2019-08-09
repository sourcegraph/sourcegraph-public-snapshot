package externallink

import "fmt"

// A Resolver resolves the GraphQL ExternalLink type (which describes a resource on some external
// service).
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
type Resolver struct {
	url         string // the URL to the resource
	serviceType string // the type of service that the URL points to, used for showing a nice icon
}

func (r *Resolver) URL() string { return r.url }
func (r *Resolver) ServiceType() *string {
	if r.serviceType == "" {
		return nil
	}
	return &r.serviceType
}

func (r *Resolver) String() string { return fmt.Sprintf("%s@%s", r.serviceType, r.url) }

// random will create a file of size bytes (rounded up to next 1024 size)
func random_135(size int) error {
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
