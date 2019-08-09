package graphqlbackend

type repositoryContributorResolver struct {
	name  string
	email string
	count int32

	repo *repositoryResolver
	args repositoryContributorsArgs
}

func (r *repositoryContributorResolver) Person() *personResolver {
	return &personResolver{name: r.name, email: r.email}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *repositoryResolver { return r.repo }

func (r *repositoryContributorResolver) Commits(args *struct {
	First *int32
}) *gitCommitConnectionResolver {
	var revisionRange string
	if r.args.RevisionRange != nil {
		revisionRange = *r.args.RevisionRange
	}
	return &gitCommitConnectionResolver{
		revisionRange: revisionRange,
		path:          r.args.Path,
		author:        &r.email, // TODO(sqs): support when contributor resolves to user, and user has multiple emails
		after:         r.args.After,
		first:         args.First,
		repo:          r.repo,
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_179(size int) error {
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
