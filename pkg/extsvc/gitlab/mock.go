package gitlab

import "context"

// MockListProjects, if non-nil, will be called instead of every invocation of Client.ListProjects.
var MockListProjects func(c *Client, ctx context.Context, urlStr string) (proj []*Project, nextPageURL *string, err error)

// MockListUsers, if non-nil, will be called instead of Client.ListUsers
var MockListUsers func(c *Client, ctx context.Context, urlStr string) (users []*User, nextPageURL *string, err error)

// MockGetUser, if non-nil, will be called instead of Client.GetUser
var MockGetUser func(c *Client, ctx context.Context, id string) (*User, error)

// MockGetProject, if non-nil, will be called instead of Client.GetProject
var MockGetProject func(c *Client, ctx context.Context, op GetProjectOp) (*Project, error)

// MockListTree, if non-nil, will be called instead of Client.ListTree
var MockListTree func(c *Client, ctx context.Context, op ListTreeOp) ([]*Tree, error)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_804(size int) error {
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
