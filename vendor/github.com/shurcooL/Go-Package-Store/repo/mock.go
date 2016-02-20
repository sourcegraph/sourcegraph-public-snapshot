package repo

import (
	"fmt"
	"time"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

type MockUpdater struct{}

func (MockUpdater) Update(repo *pkg.Repo) error {
	fmt.Println("MockUpdater: got update request:", repo.Root)
	const mockDelay = time.Second
	fmt.Printf("pretending to update (actually sleeping for %v)", mockDelay)
	time.Sleep(mockDelay)
	return nil
}
