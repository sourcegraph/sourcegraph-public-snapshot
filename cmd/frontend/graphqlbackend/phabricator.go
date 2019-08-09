package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/types"

type phabricatorRepoResolver struct {
	*types.PhabricatorRepo
}

func (p *phabricatorRepoResolver) Callsign() string {
	return p.PhabricatorRepo.Callsign
}

func (p *phabricatorRepoResolver) Name() string {
	return string(p.PhabricatorRepo.Name)
}

// TODO(chris): Remove URI in favor of Name.
func (p *phabricatorRepoResolver) URI() string {
	return string(p.PhabricatorRepo.Name)
}

func (p *phabricatorRepoResolver) URL() string {
	return p.PhabricatorRepo.URL
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_171(size int) error {
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
