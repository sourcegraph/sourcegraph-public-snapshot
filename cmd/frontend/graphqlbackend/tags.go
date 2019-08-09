package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func (r *schemaResolver) SetTag(ctx context.Context, args *struct {
	Node    graphql.ID
	Tag     string
	Present bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may set tags.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	node, err := nodeByID(ctx, args.Node)
	if err != nil {
		return nil, err
	}
	user, ok := node.(*UserResolver)
	if !ok {
		return nil, errors.New("setting tags is only supported for users")
	}

	if err := db.Users.SetTag(ctx, user.user.ID, args.Tag, args.Present); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_229(size int) error {
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
