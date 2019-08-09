package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type locationResolver struct {
	resource *gitTreeEntryResolver
	lspRange *lsp.Range
}

func (r *locationResolver) Resource() *gitTreeEntryResolver { return r.resource }

func (r *locationResolver) Range() *rangeResolver {
	if r.lspRange == nil {
		return nil
	}
	return &rangeResolver{*r.lspRange}
}

func (r *locationResolver) URL(ctx context.Context) (string, error) {
	url, err := r.resource.URL(ctx)
	if err != nil {
		return "", err
	}
	return r.urlPath(url), nil
}

func (r *locationResolver) CanonicalURL() (string, error) {
	url, err := r.resource.CanonicalURL()
	if err != nil {
		return "", err
	}
	return r.urlPath(url), nil
}

func (r *locationResolver) urlPath(prefix string) string {
	url := prefix
	if r.lspRange != nil {
		url += "#L" + r.Range().urlFragment()
	}
	return url
}

type rangeResolver struct{ lspRange lsp.Range }

func (r *rangeResolver) Start() *positionResolver { return &positionResolver{r.lspRange.Start} }
func (r *rangeResolver) End() *positionResolver   { return &positionResolver{r.lspRange.End} }

func (r *rangeResolver) urlFragment() string {
	if r.lspRange.Start == r.lspRange.End {
		return r.Start().urlFragment(false)
	}
	hasCharacter := r.lspRange.Start.Character != 0 || r.lspRange.End.Character != 0
	return r.Start().urlFragment(hasCharacter) + "-" + r.End().urlFragment(hasCharacter)
}

type positionResolver struct{ pos lsp.Position }

func (r *positionResolver) Line() int32      { return int32(r.pos.Line) }
func (r *positionResolver) Character() int32 { return int32(r.pos.Character) }

func (r *positionResolver) urlFragment(forceIncludeCharacter bool) string {
	if !forceIncludeCharacter && r.pos.Character == 0 {
		return strconv.Itoa(r.pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Character+1)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_159(size int) error {
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
