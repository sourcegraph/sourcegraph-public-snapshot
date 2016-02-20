package repo

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

// GopathUpdater is an Updater that updates Go packages in local GOPATH workspaces.
type GopathUpdater struct{}

func (GopathUpdater) Update(repo *pkg.Repo) error {
	fmt.Printf("cd %s\n", repo.Path)
	fmt.Printf("%s %s", repo.Cmd.Cmd, repo.Cmd.DownloadCmd)
	err := repo.Cmd.Download(repo.Path)
	return err
}
