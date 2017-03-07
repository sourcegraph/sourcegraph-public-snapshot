package git

import (
	"context"
	"os"

	"github.com/sourcegraph/zap/pkg/gitutil"
)

type MockGitRepo struct {
	gitRepo
	IsValidRev_                                    func(rev string) (bool, error)
	Reset_                                         func(typ, rev string) error
	Clean_                                         func() error
	Push_                                          func(remote, refspec string, force bool) error
	RemoteForBranchOrZapDefaultRemote_             func(string) (string, error)
	RemoteURL_                                     func(string) (string, error)
	UpdateSymbolicRef_                             func(name, ref string) error
	ReadSymbolicRef_                               func(name string) (string, error)
	CheckoutDetachedHEAD_                          func(ref string) error
	ReadBlob_                                      func(rev, name string) ([]byte, string, string, error)
	MakeCommit_                                    func(parent string, onlyIfChangedFiles bool) (string, []*gitutil.ChangedFile, error)
	ListTreeFull_                                  func(rev string) (*gitutil.Tree, error)
	HashObject_                                    func(typ, name string, data []byte) (string, error)
	ConfigGetOne_                                  func(name string) (string, error)
	ConfigSet_                                     func(name, value string) error
	CreateTree_                                    func(basePath string, entries []*gitutil.TreeEntry) (string, error)
	CreateCommitFromTree_                          func(tree, parent string, isRootCommit bool) (string, error)
	ObjectNameSHA_                                 func(arg string) (string, error)
	HEADHasNoCommitsAndNextCommitWillBeRootCommit_ func() (bool, error)
	HEADOrDevNullTree_                             func() (string, error)
	IsIndexLocked_                                 func() (bool, error)
}

func (m MockGitRepo) GitDir() string { return ".git" }

func (m MockGitRepo) WorktreeDir() string { return "" }

func (m MockGitRepo) IsValidRev(rev string) (bool, error) {
	return m.IsValidRev_(rev)
}

func (m MockGitRepo) Reset(typ, rev string) error {
	return m.Reset_(typ, rev)
}

func (m MockGitRepo) Clean() error {
	return m.Clean_()
}

func (m MockGitRepo) Push(remote, refspec string, force bool) error {
	return m.Push_(remote, refspec, force)
}

func (m MockGitRepo) RemoteForBranchOrZapDefaultRemote(branch string) (string, error) {
	return m.RemoteForBranchOrZapDefaultRemote_(branch)
}

func (m MockGitRepo) RemoteURL(remote string) (string, error) {
	return m.RemoteURL_(remote)
}

func (m MockGitRepo) UpdateSymbolicRef(name, ref string) error {
	return m.UpdateSymbolicRef_(name, ref)
}

func (m MockGitRepo) ReadSymbolicRef(name string) (string, error) {
	return m.ReadSymbolicRef_(name)
}

func (m MockGitRepo) CheckoutDetachedHEAD(ref string) error {
	return m.CheckoutDetachedHEAD_(ref)
}

func (m MockGitRepo) ReadBlob(rev, name string) ([]byte, string, string, error) {
	return m.ReadBlob_(rev, name)
}

func (m MockGitRepo) MakeCommit(ctx context.Context, parent string, onlyIfChangedFiles bool) (string, []*gitutil.ChangedFile, error) {
	return m.MakeCommit_(parent, onlyIfChangedFiles)
}

func (m MockGitRepo) ListTreeFull(rev string) (*gitutil.Tree, error) {
	return m.ListTreeFull_(rev)
}

func (m MockGitRepo) HashObject(typ, name string, data []byte) (string, error) {
	return m.HashObject_(typ, name, data)
}

func (m MockGitRepo) ConfigGetOne(name string) (string, error) {
	return m.ConfigGetOne_(name)
}

func (m MockGitRepo) ConfigSet(name, value string) error {
	return m.ConfigSet_(name, value)
}

func (m MockGitRepo) CreateTree(basePath string, entries []*gitutil.TreeEntry) (string, error) {
	return m.CreateTree_(basePath, entries)
}

func (m MockGitRepo) CreateCommitFromTree(ctx context.Context, tree, parent string, isRootCommit bool) (string, error) {
	return m.CreateCommitFromTree_(tree, parent, isRootCommit)
}

func (m MockGitRepo) ObjectNameSHA(arg string) (string, error) {
	return m.ObjectNameSHA_(arg)
}

func (m MockGitRepo) HEADHasNoCommitsAndNextCommitWillBeRootCommit() (bool, error) {
	return m.HEADHasNoCommitsAndNextCommitWillBeRootCommit_()
}

func (m MockGitRepo) IsIndexLocked() (bool, error) {
	return m.IsIndexLocked_()
}

func (m MockGitRepo) HEADOrDevNullTree() (string, error) {
	return m.HEADOrDevNullTree_()
}

type mockFS struct {
	FileSystem
	WriteFile_ func(name string, data []byte, mode os.FileMode) error
	Stat_      func(name string) (os.FileInfo, error)
}

func (f mockFS) WriteFile(name string, data []byte, mode os.FileMode) error {
	return f.WriteFile_(name, data, mode)
}

func (f mockFS) Stat(name string) (os.FileInfo, error) {
	return f.Stat_(name)
}
