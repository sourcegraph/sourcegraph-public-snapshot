package updater

import (
	"go/build"
	"log"
	"sync"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

type GoPackageList struct {
	// TODO: Merge the List and OrderedList into a single struct to better communicate that it's a single data structure.
	sync.Mutex
	OrderedList []*RepoPresenter          // OrderedList has the same contents as List, but gives it a stable order.
	List        map[string]*RepoPresenter // Map key is repoRoot.
}

type RepoPresenter struct {
	Repo *pkg.Repo
	presenter.Presenter
}

// workspace is a workspace environment, meaning each repo has local and remote components.
type workspace struct {
	repositories        chan Repo
	importPaths         chan string
	importPathRevisions chan importPathRevision
	unique              chan *pkg.Repo
	// processedFiltered is the output of processed repos (complete with local and remote revisions),
	// with just enough information to decide if an update should be displayed.
	processedFiltered chan *pkg.Repo
	// presented is the output of processed and presented repos (complete with repo.Presenter).
	presented chan *RepoPresenter

	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is the import path corresponding to the root of the repository.

	newObserver   chan observerRequest
	observers     map[chan *RepoPresenter]struct{}
	GoPackageList *GoPackageList
}

type observerRequest struct {
	Response chan chan *RepoPresenter
}

func NewWorkspace() *workspace {
	w := &workspace{
		importPaths:         make(chan string, 64),
		repositories:        make(chan Repo, 64),
		importPathRevisions: make(chan importPathRevision, 64),
		unique:              make(chan *pkg.Repo, 64),
		processedFiltered:   make(chan *pkg.Repo, 64),
		presented:           make(chan *RepoPresenter, 64),

		repos: make(map[string]*pkg.Repo),

		newObserver:   make(chan observerRequest),
		observers:     make(map[chan *RepoPresenter]struct{}),
		GoPackageList: &GoPackageList{List: make(map[string]*RepoPresenter)},
	}

	{
		var wg0 sync.WaitGroup
		for range iter.N(8) {
			wg0.Add(1)
			go w.importPathWorker(&wg0)
		}
		var wg1 sync.WaitGroup
		for range iter.N(8) {
			wg1.Add(1)
			go w.repositoriesWorker(&wg1)
		}
		var wg2 sync.WaitGroup
		for range iter.N(8) {
			wg2.Add(1)
			go w.importPathRevisionWorker(&wg2)
		}
		go func() {
			wg0.Wait()
			wg1.Wait()
			wg2.Wait()
			close(w.unique)
		}()
	}

	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.processFilterWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.processedFiltered)
		}()
	}

	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.presenterWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.presented)
		}()
	}

	go w.run()

	return w
}

type Repo struct {
	Path string
	Root string
	VCS  *vcs.Cmd
}

func (w *workspace) AddRepository(r Repo) {
	w.repositories <- r
}

// Add adds a package with specified import path for processing.
func (w *workspace) Add(importPath string) {
	w.importPaths <- importPath
}

type importPathRevision struct {
	importPath string
	revision   string
}

// AddRevision adds a package with specified import path and revision for processing.
func (w *workspace) AddRevision(importPath string, revision string) {
	w.importPathRevisions <- importPathRevision{
		importPath: importPath,
		revision:   revision,
	}
}

// Done should be called after the workspace is finished being populated.
func (w *workspace) Done() {
	close(w.importPaths)
	close(w.repositories)
	close(w.importPathRevisions)
}

func (w *workspace) Presented() <-chan *RepoPresenter {
	response := make(chan chan *RepoPresenter)
	w.newObserver <- observerRequest{Response: response}
	return <-response
}

func (w *workspace) run() {
Outer:
	for {
		select {
		// New repoPresenter available.
		case repoPresenter, ok := <-w.presented:
			// We're done streaming.
			if !ok {
				break Outer
			}

			// Append repoPresenter to current list.
			w.GoPackageList.Lock()
			w.GoPackageList.OrderedList = append(w.GoPackageList.OrderedList, repoPresenter)
			w.GoPackageList.List[repoPresenter.Repo.Root] = repoPresenter
			w.GoPackageList.Unlock()

			// Send new repoPresenter to all existing observers.
			for ch := range w.observers {
				ch <- repoPresenter
			}
		// New observer request.
		case req := <-w.newObserver:
			w.GoPackageList.Lock()
			ch := make(chan *RepoPresenter, len(w.GoPackageList.OrderedList))
			for _, repoPresenter := range w.GoPackageList.OrderedList {
				ch <- repoPresenter
			}
			w.GoPackageList.Unlock()

			w.observers[ch] = struct{}{}

			req.Response <- ch
		}
	}

	// At this point, streaming has finished, so finish up existing observers.
	for ch := range w.observers {
		close(ch)
	}
	w.observers = nil

	// Respond to new observer requests directly.
	for req := range w.newObserver {
		w.GoPackageList.Lock()
		ch := make(chan *RepoPresenter, len(w.GoPackageList.OrderedList))
		for _, repoPresenter := range w.GoPackageList.OrderedList {
			ch <- repoPresenter
		}
		w.GoPackageList.Unlock()

		close(ch)

		req.Response <- ch
	}
}

// repositoriesWorker sends unique repositories to phase 2.
func (w *workspace) repositoriesWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range w.repositories {
		vcsCmd, root := r.VCS, r.Root
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[root]; !ok {
			repo = &pkg.Repo{
				Path: r.Path,
				Root: root,
				Cmd:  vcsCmd,
				VCS:  vcs,
			}
			w.repos[root] = repo
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// importPathWorker sends unique repositories to phase 2.
func (w *workspace) importPathWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for importPath := range w.importPaths {
		// Determine repo root.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(importPath, "", build.FindOnly)
		if err != nil {
			log.Println("build.Import:", err)
			continue
		}
		if bpkg.Goroot {
			// Go-Package-Store has no support for updating packages in GOROOT, so skip those.
			continue
		}
		vcsCmd, root, err := vcs.FromDir(bpkg.Dir, bpkg.SrcRoot)
		if err != nil {
			// Go package not under VCS.
			continue
		}
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[root]; !ok {
			repo = &pkg.Repo{
				Path: bpkg.Dir,
				Root: root,
				Cmd:  vcsCmd,
				VCS:  vcs,
				// TODO: Maybe keep track of import paths inside, etc.
			}
			w.repos[root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// importPathRevisionWorker sends unique repositories to phase 2.
func (w *workspace) importPathRevisionWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range w.importPathRevisions {
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			log.Printf("failed to dynamically determine repo root for %v: %v\n", p.importPath, err)
			continue
		}
		remoteVCS, err := vcsstate.NewRemoteVCS(rr.VCS)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v\n", rr.Root, err)
			continue
		}

		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root:      rr.Root,
				RemoteURL: rr.Repo,
				Cmd:       rr.VCS,
				RemoteVCS: remoteVCS,
			}
			repo.Local.Revision = p.revision
			w.repos[rr.Root] = repo
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// processFilterWorker computes repository remote revision (and local if needed)
// in order to figure out if repositories should be presented.
func (w *workspace) processFilterWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range w.unique {
		// Determine remote revision.
		// This is slow because it requires a network operation.
		var remoteRevision string
		if p.VCS != nil {
			var err error
			remoteRevision, err = p.VCS.RemoteRevision(p.Path)
			_ = err // TODO.
		} else if p.RemoteVCS != nil {
			var err error
			remoteRevision, err = p.RemoteVCS.RemoteRevision(p.RemoteURL)
			_ = err // TODO.
		}

		p.Remote.Revision = remoteRevision

		// TODO: Organize.
		if p.Local.Revision == "" && p.VCS != nil {
			if r, err := p.VCS.LocalRevision(p.Path); err == nil {
				p.Local.Revision = r
			}

			// TODO: Organize.
			if p.RemoteVCS == nil && p.RemoteURL == "" {
				if r, err := p.VCS.RemoteURL(p.Path); err == nil {
					p.RemoteURL = r
				}
			}
		}

		if !shouldPresentUpdate(p) {
			continue
		}

		w.processedFiltered <- p
	}
}

// presenterWorker works with repos that should be displayed, creating a presenter each.
func (w *workspace) presenterWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for repo := range w.processedFiltered {
		// This part might take a while.
		repoPresenter := presenter.New(repo)

		w.presented <- &RepoPresenter{
			Repo:      repo,
			Presenter: repoPresenter,
		}
	}
}
