package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type Tree struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

type ListTreeOp struct {
	ProjID                int
	ProjPathWithNamespace string
	CommonOp
}

var ListTreeMock func(ctx context.Context, op ListTreeOp) ([]*Tree, error)

// ListTree lists the repository tree of the specified project. The underlying GitLab API has more
// options, but for now, we only support non-recursive queries of the root directory. Requests
// results are not cached by the client at the moment (i.e., setting op.NoCache to true does not
// alter behavior).
func (c *Client) ListTree(ctx context.Context, op ListTreeOp) ([]*Tree, error) {
	if MockListTree != nil {
		return MockListTree(c, ctx, op)
	}

	if op.ProjID != 0 && op.ProjPathWithNamespace != "" {
		return nil, errors.New("invalid args (specify exactly one of id and projPathWithNamespace")
	}

	return c.listTreeFromAPI(ctx, op.ProjID, op.ProjPathWithNamespace)
}

func (c *Client) listTreeFromAPI(ctx context.Context, projID int, projPathWithNamespace string) (tree []*Tree, err error) {
	var projSpecifier string
	if projID != 0 {
		projSpecifier = strconv.Itoa(projID)
	} else {
		projSpecifier = url.PathEscape(projPathWithNamespace)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%s/repository/tree", projSpecifier), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.do(ctx, req, &tree)
	return tree, err
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_807(size int) error {
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
