// Package repo provides an Updater interface and implementations.
package repo

import "github.com/shurcooL/Go-Package-Store/pkg"

// Updater is able to update Go packages contained in repositories.
type Updater interface {
	// Update specified repository to latest version.
	Update(repo *pkg.Repo) error
}
