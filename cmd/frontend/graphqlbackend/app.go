package graphqlbackend

import (
	"context"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/internal/singleprogram/filepicker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type LocalDirectoryArgs struct {
	Dir string
}

type AppResolver interface {
	LocalDirectoryPicker(ctx context.Context) (LocalDirectoryResolver, error)
	LocalDirectory(ctx context.Context, args *LocalDirectoryArgs) (LocalDirectoryResolver, error)
}

type LocalDirectoryResolver interface {
	Path() string
	Repositories() ([]LocalRepositoryResolver, error)
}

type LocalRepositoryResolver interface {
	Name() string
	Path() string
}

type appResolver struct {
	logger log.Logger
	db     database.DB
}

var _ AppResolver = &appResolver{}

func NewAppResolver(logger log.Logger, db database.DB) *appResolver {
	return &appResolver{
		logger: logger,
		db:     db,
	}
}

func (r *appResolver) checkLocalDirectoryAccess(ctx context.Context) error {
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *appResolver) LocalDirectoryPicker(ctx context.Context) (LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	picker, ok := filepicker.Lookup(r.logger)
	if !ok {
		return nil, errors.New("filepicker is not available")
	}

	path, err := picker(ctx)
	if err != nil {
		return nil, err
	}

	return &localDirectoryResolver{path: path}, nil
}

func (r *appResolver) LocalDirectory(ctx context.Context, args *LocalDirectoryArgs) (LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	path, err := filepath.Abs(args.Dir)
	if err != nil {
		return nil, err
	}

	return &localDirectoryResolver{path: path}, nil
}

type localDirectoryResolver struct {
	path string
}

func (r *localDirectoryResolver) Path() string {
	return r.path
}

func (r *localDirectoryResolver) Repositories() ([]LocalRepositoryResolver, error) {
	var c servegit.ServeConfig
	c.Load()

	srv := &servegit.Serve{
		ServeConfig: c,
		Logger:      log.Scoped("serve", ""),
	}

	repos, err := srv.Repos(r.path)
	if err != nil {
		return nil, err
	}

	local := make([]LocalRepositoryResolver, 0, len(repos))
	for _, repo := range repos {
		local = append(local, localRepositoryResolver{
			name: repo.Name,
			path: filepath.Join(r.path, repo.Name), // TODO(keegan) this is not always correct
		})
	}

	return local, nil
}

type localRepositoryResolver struct {
	name string
	path string
}

func (r localRepositoryResolver) Name() string {
	return r.name
}

func (r localRepositoryResolver) Path() string {
	return r.path
}
