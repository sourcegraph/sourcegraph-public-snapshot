// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge gitserver

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"sync"
	"time"

	diff "github.com/sourcegrbph/go-diff/diff"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	buthz "github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	gitdombin "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	protocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
)

// MockClient is b mock implementbtion of the Client interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/gitserver) used for
// unit testing.
type MockClient struct {
	// AddrForRepoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method AddrForRepo.
	AddrForRepoFunc *ClientAddrForRepoFunc
	// AddrsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Addrs.
	AddrsFunc *ClientAddrsFunc
	// ArchiveRebderFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ArchiveRebder.
	ArchiveRebderFunc *ClientArchiveRebderFunc
	// BbtchLogFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method BbtchLog.
	BbtchLogFunc *ClientBbtchLogFunc
	// BlbmeFileFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method BlbmeFile.
	BlbmeFileFunc *ClientBlbmeFileFunc
	// BrbnchesContbiningFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BrbnchesContbining.
	BrbnchesContbiningFunc *ClientBrbnchesContbiningFunc
	// CommitDbteFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitDbte.
	CommitDbteFunc *ClientCommitDbteFunc
	// CommitExistsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitExists.
	CommitExistsFunc *ClientCommitExistsFunc
	// CommitGrbphFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitGrbph.
	CommitGrbphFunc *ClientCommitGrbphFunc
	// CommitLogFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitLog.
	CommitLogFunc *ClientCommitLogFunc
	// CommitsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Commits.
	CommitsFunc *ClientCommitsFunc
	// CommitsExistFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CommitsExist.
	CommitsExistFunc *ClientCommitsExistFunc
	// CommitsUniqueToBrbnchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CommitsUniqueToBrbnch.
	CommitsUniqueToBrbnchFunc *ClientCommitsUniqueToBrbnchFunc
	// ContributorCountFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ContributorCount.
	ContributorCountFunc *ClientContributorCountFunc
	// CrebteCommitFromPbtchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteCommitFromPbtch.
	CrebteCommitFromPbtchFunc *ClientCrebteCommitFromPbtchFunc
	// DiffFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Diff.
	DiffFunc *ClientDiffFunc
	// DiffPbthFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method DiffPbth.
	DiffPbthFunc *ClientDiffPbthFunc
	// DiffSymbolsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DiffSymbols.
	DiffSymbolsFunc *ClientDiffSymbolsFunc
	// FirstEverCommitFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method FirstEverCommit.
	FirstEverCommitFunc *ClientFirstEverCommitFunc
	// GetBehindAhebdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBehindAhebd.
	GetBehindAhebdFunc *ClientGetBehindAhebdFunc
	// GetCommitFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetCommit.
	GetCommitFunc *ClientGetCommitFunc
	// GetCommitsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetCommits.
	GetCommitsFunc *ClientGetCommitsFunc
	// GetDefbultBrbnchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDefbultBrbnch.
	GetDefbultBrbnchFunc *ClientGetDefbultBrbnchFunc
	// GetObjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetObject.
	GetObjectFunc *ClientGetObjectFunc
	// HbsCommitAfterFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method HbsCommitAfter.
	HbsCommitAfterFunc *ClientHbsCommitAfterFunc
	// HebdFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hebd.
	HebdFunc *ClientHebdFunc
	// IsRepoClonebbleFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method IsRepoClonebble.
	IsRepoClonebbleFunc *ClientIsRepoClonebbleFunc
	// ListBrbnchesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ListBrbnches.
	ListBrbnchesFunc *ClientListBrbnchesFunc
	// ListDirectoryChildrenFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListDirectoryChildren.
	ListDirectoryChildrenFunc *ClientListDirectoryChildrenFunc
	// ListRefsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method ListRefs.
	ListRefsFunc *ClientListRefsFunc
	// ListTbgsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method ListTbgs.
	ListTbgsFunc *ClientListTbgsFunc
	// LogReverseEbchFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LogReverseEbch.
	LogReverseEbchFunc *ClientLogReverseEbchFunc
	// LsFilesFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LsFiles.
	LsFilesFunc *ClientLsFilesFunc
	// MergeBbseFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MergeBbse.
	MergeBbseFunc *ClientMergeBbseFunc
	// NewFileRebderFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewFileRebder.
	NewFileRebderFunc *ClientNewFileRebderFunc
	// P4ExecFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method P4Exec.
	P4ExecFunc *ClientP4ExecFunc
	// P4GetChbngelistFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method P4GetChbngelist.
	P4GetChbngelistFunc *ClientP4GetChbngelistFunc
	// RebdDirFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RebdDir.
	RebdDirFunc *ClientRebdDirFunc
	// RebdFileFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RebdFile.
	RebdFileFunc *ClientRebdFileFunc
	// RefDescriptionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RefDescriptions.
	RefDescriptionsFunc *ClientRefDescriptionsFunc
	// RemoveFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Remove.
	RemoveFunc *ClientRemoveFunc
	// RepoCloneProgressFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepoCloneProgress.
	RepoCloneProgressFunc *ClientRepoCloneProgressFunc
	// RequestRepoCloneFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RequestRepoClone.
	RequestRepoCloneFunc *ClientRequestRepoCloneFunc
	// RequestRepoUpdbteFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RequestRepoUpdbte.
	RequestRepoUpdbteFunc *ClientRequestRepoUpdbteFunc
	// ResolveRevisionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveRevision.
	ResolveRevisionFunc *ClientResolveRevisionFunc
	// ResolveRevisionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveRevisions.
	ResolveRevisionsFunc *ClientResolveRevisionsFunc
	// RevListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method RevList.
	RevListFunc *ClientRevListFunc
	// SebrchFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Sebrch.
	SebrchFunc *ClientSebrchFunc
	// StbtFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Stbt.
	StbtFunc *ClientStbtFunc
	// StrebmBlbmeFileFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method StrebmBlbmeFile.
	StrebmBlbmeFileFunc *ClientStrebmBlbmeFileFunc
	// SystemInfoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SystemInfo.
	SystemInfoFunc *ClientSystemInfoFunc
	// SystemsInfoFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SystemsInfo.
	SystemsInfoFunc *ClientSystemsInfoFunc
}

// NewMockClient crebtes b new mock of the Client interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockClient() *MockClient {
	return &MockClient{
		AddrForRepoFunc: &ClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 string) {
				return
			},
		},
		AddrsFunc: &ClientAddrsFunc{
			defbultHook: func() (r0 []string) {
				return
			},
		},
		ArchiveRebderFunc: &ClientArchiveRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		BbtchLogFunc: &ClientBbtchLogFunc{
			defbultHook: func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) (r0 error) {
				return
			},
		},
		BlbmeFileFunc: &ClientBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (r0 []*Hunk, r1 error) {
				return
			},
		},
		BrbnchesContbiningFunc: &ClientBrbnchesContbiningFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 []string, r1 error) {
				return
			},
		},
		CommitDbteFunc: &ClientCommitDbteFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 string, r1 time.Time, r2 bool, r3 error) {
				return
			},
		},
		CommitExistsFunc: &ClientCommitExistsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (r0 bool, r1 error) {
				return
			},
		},
		CommitGrbphFunc: &ClientCommitGrbphFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (r0 *gitdombin.CommitGrbph, r1 error) {
				return
			},
		},
		CommitLogFunc: &ClientCommitLogFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Time) (r0 []CommitLog, r1 error) {
				return
			},
		},
		CommitsFunc: &ClientCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) (r0 []*gitdombin.Commit, r1 error) {
				return
			},
		},
		CommitsExistFunc: &ClientCommitsExistFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) (r0 []bool, r1 error) {
				return
			},
		},
		CommitsUniqueToBrbnchFunc: &ClientCommitsUniqueToBrbnchFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (r0 mbp[string]time.Time, r1 error) {
				return
			},
		},
		ContributorCountFunc: &ClientContributorCountFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ContributorOptions) (r0 []*gitdombin.ContributorCount, r1 error) {
				return
			},
		},
		CrebteCommitFromPbtchFunc: &ClientCrebteCommitFromPbtchFunc{
			defbultHook: func(context.Context, protocol.CrebteCommitFromPbtchRequest) (r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
				return
			},
		},
		DiffFunc: &ClientDiffFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (r0 *DiffFileIterbtor, r1 error) {
				return
			},
		},
		DiffPbthFunc: &ClientDiffPbthFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) (r0 []*diff.Hunk, r1 error) {
				return
			},
		},
		DiffSymbolsFunc: &ClientDiffSymbolsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (r0 []byte, r1 error) {
				return
			},
		},
		FirstEverCommitFunc: &ClientFirstEverCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (r0 *gitdombin.Commit, r1 error) {
				return
			},
		},
		GetBehindAhebdFunc: &ClientGetBehindAhebdFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (r0 *gitdombin.BehindAhebd, r1 error) {
				return
			},
		},
		GetCommitFunc: &ClientGetCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (r0 *gitdombin.Commit, r1 error) {
				return
			},
		},
		GetCommitsFunc: &ClientGetCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) (r0 []*gitdombin.Commit, r1 error) {
				return
			},
		},
		GetDefbultBrbnchFunc: &ClientGetDefbultBrbnchFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bool) (r0 string, r1 bpi.CommitID, r2 error) {
				return
			},
		},
		GetObjectFunc: &ClientGetObjectFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string) (r0 *gitdombin.GitObject, r1 error) {
				return
			},
		},
		HbsCommitAfterFunc: &ClientHbsCommitAfterFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (r0 bool, r1 error) {
				return
			},
		},
		HebdFunc: &ClientHebdFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (r0 string, r1 bool, r2 error) {
				return
			},
		},
		IsRepoClonebbleFunc: &ClientIsRepoClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 error) {
				return
			},
		},
		ListBrbnchesFunc: &ClientListBrbnchesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, BrbnchesOptions) (r0 []*gitdombin.Brbnch, r1 error) {
				return
			},
		},
		ListDirectoryChildrenFunc: &ClientListDirectoryChildrenFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (r0 mbp[string][]string, r1 error) {
				return
			},
		},
		ListRefsFunc: &ClientListRefsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 []gitdombin.Ref, r1 error) {
				return
			},
		},
		ListTbgsFunc: &ClientListTbgsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ...string) (r0 []*gitdombin.Tbg, r1 error) {
				return
			},
		},
		LogReverseEbchFunc: &ClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) (r0 error) {
				return
			},
		},
		LsFilesFunc: &ClientLsFilesFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) (r0 []string, r1 error) {
				return
			},
		},
		MergeBbseFunc: &ClientMergeBbseFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (r0 bpi.CommitID, r1 error) {
				return
			},
		},
		NewFileRebderFunc: &ClientNewFileRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		P4ExecFunc: &ClientP4ExecFunc{
			defbultHook: func(context.Context, string, string, string, ...string) (r0 io.RebdCloser, r1 http.Hebder, r2 error) {
				return
			},
		},
		P4GetChbngelistFunc: &ClientP4GetChbngelistFunc{
			defbultHook: func(context.Context, string, PerforceCredentibls) (r0 *protocol.PerforceChbngelist, r1 error) {
				return
			},
		},
		RebdDirFunc: &ClientRebdDirFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) (r0 []fs.FileInfo, r1 error) {
				return
			},
		},
		RebdFileFunc: &ClientRebdFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 []byte, r1 error) {
				return
			},
		},
		RefDescriptionsFunc: &ClientRefDescriptionsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (r0 mbp[string][]gitdombin.RefDescription, r1 error) {
				return
			},
		},
		RemoveFunc: &ClientRemoveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 error) {
				return
			},
		},
		RepoCloneProgressFunc: &ClientRepoCloneProgressFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (r0 *protocol.RepoCloneProgressResponse, r1 error) {
				return
			},
		},
		RequestRepoCloneFunc: &ClientRequestRepoCloneFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 *protocol.RepoCloneResponse, r1 error) {
				return
			},
		},
		RequestRepoUpdbteFunc: &ClientRequestRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Durbtion) (r0 *protocol.RepoUpdbteResponse, r1 error) {
				return
			},
		},
		ResolveRevisionFunc: &ClientResolveRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (r0 bpi.CommitID, r1 error) {
				return
			},
		},
		ResolveRevisionsFunc: &ClientResolveRevisionsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) (r0 []string, r1 error) {
				return
			},
		},
		RevListFunc: &ClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) (r0 error) {
				return
			},
		},
		SebrchFunc: &ClientSebrchFunc{
			defbultHook: func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (r0 bool, r1 error) {
				return
			},
		},
		StbtFunc: &ClientStbtFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (r0 fs.FileInfo, r1 error) {
				return
			},
		},
		StrebmBlbmeFileFunc: &ClientStrebmBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (r0 HunkRebder, r1 error) {
				return
			},
		},
		SystemInfoFunc: &ClientSystemInfoFunc{
			defbultHook: func(context.Context, string) (r0 SystemInfo, r1 error) {
				return
			},
		},
		SystemsInfoFunc: &ClientSystemsInfoFunc{
			defbultHook: func(context.Context) (r0 []SystemInfo, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockClient crebtes b new mock of the Client interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockClient() *MockClient {
	return &MockClient{
		AddrForRepoFunc: &ClientAddrForRepoFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) string {
				pbnic("unexpected invocbtion of MockClient.AddrForRepo")
			},
		},
		AddrsFunc: &ClientAddrsFunc{
			defbultHook: func() []string {
				pbnic("unexpected invocbtion of MockClient.Addrs")
			},
		},
		ArchiveRebderFunc: &ClientArchiveRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockClient.ArchiveRebder")
			},
		},
		BbtchLogFunc: &ClientBbtchLogFunc{
			defbultHook: func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error {
				pbnic("unexpected invocbtion of MockClient.BbtchLog")
			},
		},
		BlbmeFileFunc: &ClientBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error) {
				pbnic("unexpected invocbtion of MockClient.BlbmeFile")
			},
		},
		BrbnchesContbiningFunc: &ClientBrbnchesContbiningFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
				pbnic("unexpected invocbtion of MockClient.BrbnchesContbining")
			},
		},
		CommitDbteFunc: &ClientCommitDbteFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
				pbnic("unexpected invocbtion of MockClient.CommitDbte")
			},
		},
		CommitExistsFunc: &ClientCommitExistsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
				pbnic("unexpected invocbtion of MockClient.CommitExists")
			},
		},
		CommitGrbphFunc: &ClientCommitGrbphFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
				pbnic("unexpected invocbtion of MockClient.CommitGrbph")
			},
		},
		CommitLogFunc: &ClientCommitLogFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error) {
				pbnic("unexpected invocbtion of MockClient.CommitLog")
			},
		},
		CommitsFunc: &ClientCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockClient.Commits")
			},
		},
		CommitsExistFunc: &ClientCommitsExistFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
				pbnic("unexpected invocbtion of MockClient.CommitsExist")
			},
		},
		CommitsUniqueToBrbnchFunc: &ClientCommitsUniqueToBrbnchFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
				pbnic("unexpected invocbtion of MockClient.CommitsUniqueToBrbnch")
			},
		},
		ContributorCountFunc: &ClientContributorCountFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error) {
				pbnic("unexpected invocbtion of MockClient.ContributorCount")
			},
		},
		CrebteCommitFromPbtchFunc: &ClientCrebteCommitFromPbtchFunc{
			defbultHook: func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
				pbnic("unexpected invocbtion of MockClient.CrebteCommitFromPbtch")
			},
		},
		DiffFunc: &ClientDiffFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error) {
				pbnic("unexpected invocbtion of MockClient.Diff")
			},
		},
		DiffPbthFunc: &ClientDiffPbthFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
				pbnic("unexpected invocbtion of MockClient.DiffPbth")
			},
		},
		DiffSymbolsFunc: &ClientDiffSymbolsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
				pbnic("unexpected invocbtion of MockClient.DiffSymbols")
			},
		},
		FirstEverCommitFunc: &ClientFirstEverCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockClient.FirstEverCommit")
			},
		},
		GetBehindAhebdFunc: &ClientGetBehindAhebdFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
				pbnic("unexpected invocbtion of MockClient.GetBehindAhebd")
			},
		},
		GetCommitFunc: &ClientGetCommitFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockClient.GetCommit")
			},
		},
		GetCommitsFunc: &ClientGetCommitsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
				pbnic("unexpected invocbtion of MockClient.GetCommits")
			},
		},
		GetDefbultBrbnchFunc: &ClientGetDefbultBrbnchFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockClient.GetDefbultBrbnch")
			},
		},
		GetObjectFunc: &ClientGetObjectFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
				pbnic("unexpected invocbtion of MockClient.GetObject")
			},
		},
		HbsCommitAfterFunc: &ClientHbsCommitAfterFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
				pbnic("unexpected invocbtion of MockClient.HbsCommitAfter")
			},
		},
		HebdFunc: &ClientHebdFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
				pbnic("unexpected invocbtion of MockClient.Hebd")
			},
		},
		IsRepoClonebbleFunc: &ClientIsRepoClonebbleFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) error {
				pbnic("unexpected invocbtion of MockClient.IsRepoClonebble")
			},
		},
		ListBrbnchesFunc: &ClientListBrbnchesFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
				pbnic("unexpected invocbtion of MockClient.ListBrbnches")
			},
		},
		ListDirectoryChildrenFunc: &ClientListDirectoryChildrenFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
				pbnic("unexpected invocbtion of MockClient.ListDirectoryChildren")
			},
		},
		ListRefsFunc: &ClientListRefsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
				pbnic("unexpected invocbtion of MockClient.ListRefs")
			},
		},
		ListTbgsFunc: &ClientListTbgsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
				pbnic("unexpected invocbtion of MockClient.ListTbgs")
			},
		},
		LogReverseEbchFunc: &ClientLogReverseEbchFunc{
			defbultHook: func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
				pbnic("unexpected invocbtion of MockClient.LogReverseEbch")
			},
		},
		LsFilesFunc: &ClientLsFilesFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
				pbnic("unexpected invocbtion of MockClient.LsFiles")
			},
		},
		MergeBbseFunc: &ClientMergeBbseFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockClient.MergeBbse")
			},
		},
		NewFileRebderFunc: &ClientNewFileRebderFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockClient.NewFileRebder")
			},
		},
		P4ExecFunc: &ClientP4ExecFunc{
			defbultHook: func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
				pbnic("unexpected invocbtion of MockClient.P4Exec")
			},
		},
		P4GetChbngelistFunc: &ClientP4GetChbngelistFunc{
			defbultHook: func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
				pbnic("unexpected invocbtion of MockClient.P4GetChbngelist")
			},
		},
		RebdDirFunc: &ClientRebdDirFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
				pbnic("unexpected invocbtion of MockClient.RebdDir")
			},
		},
		RebdFileFunc: &ClientRebdFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
				pbnic("unexpected invocbtion of MockClient.RebdFile")
			},
		},
		RefDescriptionsFunc: &ClientRefDescriptionsFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
				pbnic("unexpected invocbtion of MockClient.RefDescriptions")
			},
		},
		RemoveFunc: &ClientRemoveFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) error {
				pbnic("unexpected invocbtion of MockClient.Remove")
			},
		},
		RepoCloneProgressFunc: &ClientRepoCloneProgressFunc{
			defbultHook: func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
				pbnic("unexpected invocbtion of MockClient.RepoCloneProgress")
			},
		},
		RequestRepoCloneFunc: &ClientRequestRepoCloneFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
				pbnic("unexpected invocbtion of MockClient.RequestRepoClone")
			},
		},
		RequestRepoUpdbteFunc: &ClientRequestRepoUpdbteFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
				pbnic("unexpected invocbtion of MockClient.RequestRepoUpdbte")
			},
		},
		ResolveRevisionFunc: &ClientResolveRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error) {
				pbnic("unexpected invocbtion of MockClient.ResolveRevision")
			},
		},
		ResolveRevisionsFunc: &ClientResolveRevisionsFunc{
			defbultHook: func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
				pbnic("unexpected invocbtion of MockClient.ResolveRevisions")
			},
		},
		RevListFunc: &ClientRevListFunc{
			defbultHook: func(context.Context, string, string, func(commit string) (bool, error)) error {
				pbnic("unexpected invocbtion of MockClient.RevList")
			},
		},
		SebrchFunc: &ClientSebrchFunc{
			defbultHook: func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
				pbnic("unexpected invocbtion of MockClient.Sebrch")
			},
		},
		StbtFunc: &ClientStbtFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
				pbnic("unexpected invocbtion of MockClient.Stbt")
			},
		},
		StrebmBlbmeFileFunc: &ClientStrebmBlbmeFileFunc{
			defbultHook: func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error) {
				pbnic("unexpected invocbtion of MockClient.StrebmBlbmeFile")
			},
		},
		SystemInfoFunc: &ClientSystemInfoFunc{
			defbultHook: func(context.Context, string) (SystemInfo, error) {
				pbnic("unexpected invocbtion of MockClient.SystemInfo")
			},
		},
		SystemsInfoFunc: &ClientSystemsInfoFunc{
			defbultHook: func(context.Context) ([]SystemInfo, error) {
				pbnic("unexpected invocbtion of MockClient.SystemsInfo")
			},
		},
	}
}

// NewMockClientFrom crebtes b new mock of the MockClient interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockClientFrom(i Client) *MockClient {
	return &MockClient{
		AddrForRepoFunc: &ClientAddrForRepoFunc{
			defbultHook: i.AddrForRepo,
		},
		AddrsFunc: &ClientAddrsFunc{
			defbultHook: i.Addrs,
		},
		ArchiveRebderFunc: &ClientArchiveRebderFunc{
			defbultHook: i.ArchiveRebder,
		},
		BbtchLogFunc: &ClientBbtchLogFunc{
			defbultHook: i.BbtchLog,
		},
		BlbmeFileFunc: &ClientBlbmeFileFunc{
			defbultHook: i.BlbmeFile,
		},
		BrbnchesContbiningFunc: &ClientBrbnchesContbiningFunc{
			defbultHook: i.BrbnchesContbining,
		},
		CommitDbteFunc: &ClientCommitDbteFunc{
			defbultHook: i.CommitDbte,
		},
		CommitExistsFunc: &ClientCommitExistsFunc{
			defbultHook: i.CommitExists,
		},
		CommitGrbphFunc: &ClientCommitGrbphFunc{
			defbultHook: i.CommitGrbph,
		},
		CommitLogFunc: &ClientCommitLogFunc{
			defbultHook: i.CommitLog,
		},
		CommitsFunc: &ClientCommitsFunc{
			defbultHook: i.Commits,
		},
		CommitsExistFunc: &ClientCommitsExistFunc{
			defbultHook: i.CommitsExist,
		},
		CommitsUniqueToBrbnchFunc: &ClientCommitsUniqueToBrbnchFunc{
			defbultHook: i.CommitsUniqueToBrbnch,
		},
		ContributorCountFunc: &ClientContributorCountFunc{
			defbultHook: i.ContributorCount,
		},
		CrebteCommitFromPbtchFunc: &ClientCrebteCommitFromPbtchFunc{
			defbultHook: i.CrebteCommitFromPbtch,
		},
		DiffFunc: &ClientDiffFunc{
			defbultHook: i.Diff,
		},
		DiffPbthFunc: &ClientDiffPbthFunc{
			defbultHook: i.DiffPbth,
		},
		DiffSymbolsFunc: &ClientDiffSymbolsFunc{
			defbultHook: i.DiffSymbols,
		},
		FirstEverCommitFunc: &ClientFirstEverCommitFunc{
			defbultHook: i.FirstEverCommit,
		},
		GetBehindAhebdFunc: &ClientGetBehindAhebdFunc{
			defbultHook: i.GetBehindAhebd,
		},
		GetCommitFunc: &ClientGetCommitFunc{
			defbultHook: i.GetCommit,
		},
		GetCommitsFunc: &ClientGetCommitsFunc{
			defbultHook: i.GetCommits,
		},
		GetDefbultBrbnchFunc: &ClientGetDefbultBrbnchFunc{
			defbultHook: i.GetDefbultBrbnch,
		},
		GetObjectFunc: &ClientGetObjectFunc{
			defbultHook: i.GetObject,
		},
		HbsCommitAfterFunc: &ClientHbsCommitAfterFunc{
			defbultHook: i.HbsCommitAfter,
		},
		HebdFunc: &ClientHebdFunc{
			defbultHook: i.Hebd,
		},
		IsRepoClonebbleFunc: &ClientIsRepoClonebbleFunc{
			defbultHook: i.IsRepoClonebble,
		},
		ListBrbnchesFunc: &ClientListBrbnchesFunc{
			defbultHook: i.ListBrbnches,
		},
		ListDirectoryChildrenFunc: &ClientListDirectoryChildrenFunc{
			defbultHook: i.ListDirectoryChildren,
		},
		ListRefsFunc: &ClientListRefsFunc{
			defbultHook: i.ListRefs,
		},
		ListTbgsFunc: &ClientListTbgsFunc{
			defbultHook: i.ListTbgs,
		},
		LogReverseEbchFunc: &ClientLogReverseEbchFunc{
			defbultHook: i.LogReverseEbch,
		},
		LsFilesFunc: &ClientLsFilesFunc{
			defbultHook: i.LsFiles,
		},
		MergeBbseFunc: &ClientMergeBbseFunc{
			defbultHook: i.MergeBbse,
		},
		NewFileRebderFunc: &ClientNewFileRebderFunc{
			defbultHook: i.NewFileRebder,
		},
		P4ExecFunc: &ClientP4ExecFunc{
			defbultHook: i.P4Exec,
		},
		P4GetChbngelistFunc: &ClientP4GetChbngelistFunc{
			defbultHook: i.P4GetChbngelist,
		},
		RebdDirFunc: &ClientRebdDirFunc{
			defbultHook: i.RebdDir,
		},
		RebdFileFunc: &ClientRebdFileFunc{
			defbultHook: i.RebdFile,
		},
		RefDescriptionsFunc: &ClientRefDescriptionsFunc{
			defbultHook: i.RefDescriptions,
		},
		RemoveFunc: &ClientRemoveFunc{
			defbultHook: i.Remove,
		},
		RepoCloneProgressFunc: &ClientRepoCloneProgressFunc{
			defbultHook: i.RepoCloneProgress,
		},
		RequestRepoCloneFunc: &ClientRequestRepoCloneFunc{
			defbultHook: i.RequestRepoClone,
		},
		RequestRepoUpdbteFunc: &ClientRequestRepoUpdbteFunc{
			defbultHook: i.RequestRepoUpdbte,
		},
		ResolveRevisionFunc: &ClientResolveRevisionFunc{
			defbultHook: i.ResolveRevision,
		},
		ResolveRevisionsFunc: &ClientResolveRevisionsFunc{
			defbultHook: i.ResolveRevisions,
		},
		RevListFunc: &ClientRevListFunc{
			defbultHook: i.RevList,
		},
		SebrchFunc: &ClientSebrchFunc{
			defbultHook: i.Sebrch,
		},
		StbtFunc: &ClientStbtFunc{
			defbultHook: i.Stbt,
		},
		StrebmBlbmeFileFunc: &ClientStrebmBlbmeFileFunc{
			defbultHook: i.StrebmBlbmeFile,
		},
		SystemInfoFunc: &ClientSystemInfoFunc{
			defbultHook: i.SystemInfo,
		},
		SystemsInfoFunc: &ClientSystemsInfoFunc{
			defbultHook: i.SystemsInfo,
		},
	}
}

// ClientAddrForRepoFunc describes the behbvior when the AddrForRepo method
// of the pbrent MockClient instbnce is invoked.
type ClientAddrForRepoFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) string
	hooks       []func(context.Context, bpi.RepoNbme) string
	history     []ClientAddrForRepoFuncCbll
	mutex       sync.Mutex
}

// AddrForRepo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) AddrForRepo(v0 context.Context, v1 bpi.RepoNbme) string {
	r0 := m.AddrForRepoFunc.nextHook()(v0, v1)
	m.AddrForRepoFunc.bppendCbll(ClientAddrForRepoFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the AddrForRepo method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientAddrForRepoFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddrForRepo method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientAddrForRepoFunc) PushHook(hook func(context.Context, bpi.RepoNbme) string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientAddrForRepoFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientAddrForRepoFunc) PushReturn(r0 string) {
	f.PushHook(func(context.Context, bpi.RepoNbme) string {
		return r0
	})
}

func (f *ClientAddrForRepoFunc) nextHook() func(context.Context, bpi.RepoNbme) string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientAddrForRepoFunc) bppendCbll(r0 ClientAddrForRepoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientAddrForRepoFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientAddrForRepoFunc) History() []ClientAddrForRepoFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientAddrForRepoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientAddrForRepoFuncCbll is bn object thbt describes bn invocbtion of
// method AddrForRepo on bn instbnce of MockClient.
type ClientAddrForRepoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientAddrForRepoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientAddrForRepoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientAddrsFunc describes the behbvior when the Addrs method of the
// pbrent MockClient instbnce is invoked.
type ClientAddrsFunc struct {
	defbultHook func() []string
	hooks       []func() []string
	history     []ClientAddrsFuncCbll
	mutex       sync.Mutex
}

// Addrs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Addrs() []string {
	r0 := m.AddrsFunc.nextHook()()
	m.AddrsFunc.bppendCbll(ClientAddrsFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Addrs method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientAddrsFunc) SetDefbultHook(hook func() []string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Addrs method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientAddrsFunc) PushHook(hook func() []string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientAddrsFunc) SetDefbultReturn(r0 []string) {
	f.SetDefbultHook(func() []string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientAddrsFunc) PushReturn(r0 []string) {
	f.PushHook(func() []string {
		return r0
	})
}

func (f *ClientAddrsFunc) nextHook() func() []string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientAddrsFunc) bppendCbll(r0 ClientAddrsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientAddrsFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientAddrsFunc) History() []ClientAddrsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientAddrsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientAddrsFuncCbll is bn object thbt describes bn invocbtion of method
// Addrs on bn instbnce of MockClient.
type ClientAddrsFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientAddrsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientAddrsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientArchiveRebderFunc describes the behbvior when the ArchiveRebder
// method of the pbrent MockClient instbnce is invoked.
type ClientArchiveRebderFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error)
	history     []ClientArchiveRebderFuncCbll
	mutex       sync.Mutex
}

// ArchiveRebder delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ArchiveRebder(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 ArchiveOptions) (io.RebdCloser, error) {
	r0, r1 := m.ArchiveRebderFunc.nextHook()(v0, v1, v2, v3)
	m.ArchiveRebderFunc.bppendCbll(ClientArchiveRebderFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ArchiveRebder method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientArchiveRebderFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ArchiveRebder method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientArchiveRebderFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientArchiveRebderFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientArchiveRebderFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *ClientArchiveRebderFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ArchiveOptions) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientArchiveRebderFunc) bppendCbll(r0 ClientArchiveRebderFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientArchiveRebderFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientArchiveRebderFunc) History() []ClientArchiveRebderFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientArchiveRebderFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientArchiveRebderFuncCbll is bn object thbt describes bn invocbtion of
// method ArchiveRebder on bn instbnce of MockClient.
type ClientArchiveRebderFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 ArchiveOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientArchiveRebderFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientArchiveRebderFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientBbtchLogFunc describes the behbvior when the BbtchLog method of the
// pbrent MockClient instbnce is invoked.
type ClientBbtchLogFunc struct {
	defbultHook func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error
	hooks       []func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error
	history     []ClientBbtchLogFuncCbll
	mutex       sync.Mutex
}

// BbtchLog delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) BbtchLog(v0 context.Context, v1 BbtchLogOptions, v2 BbtchLogCbllbbck) error {
	r0 := m.BbtchLogFunc.nextHook()(v0, v1, v2)
	m.BbtchLogFunc.bppendCbll(ClientBbtchLogFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the BbtchLog method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientBbtchLogFunc) SetDefbultHook(hook func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BbtchLog method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientBbtchLogFunc) PushHook(hook func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientBbtchLogFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientBbtchLogFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error {
		return r0
	})
}

func (f *ClientBbtchLogFunc) nextHook() func(context.Context, BbtchLogOptions, BbtchLogCbllbbck) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientBbtchLogFunc) bppendCbll(r0 ClientBbtchLogFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientBbtchLogFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientBbtchLogFunc) History() []ClientBbtchLogFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientBbtchLogFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientBbtchLogFuncCbll is bn object thbt describes bn invocbtion of
// method BbtchLog on bn instbnce of MockClient.
type ClientBbtchLogFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 BbtchLogOptions
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 BbtchLogCbllbbck
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientBbtchLogFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientBbtchLogFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientBlbmeFileFunc describes the behbvior when the BlbmeFile method of
// the pbrent MockClient instbnce is invoked.
type ClientBlbmeFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error)
	history     []ClientBlbmeFileFuncCbll
	mutex       sync.Mutex
}

// BlbmeFile delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) BlbmeFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 *BlbmeOptions) ([]*Hunk, error) {
	r0, r1 := m.BlbmeFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.BlbmeFileFunc.bppendCbll(ClientBlbmeFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the BlbmeFile method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientBlbmeFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BlbmeFile method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientBlbmeFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientBlbmeFileFunc) SetDefbultReturn(r0 []*Hunk, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientBlbmeFileFunc) PushReturn(r0 []*Hunk, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error) {
		return r0, r1
	})
}

func (f *ClientBlbmeFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) ([]*Hunk, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientBlbmeFileFunc) bppendCbll(r0 ClientBlbmeFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientBlbmeFileFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientBlbmeFileFunc) History() []ClientBlbmeFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientBlbmeFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientBlbmeFileFuncCbll is bn object thbt describes bn invocbtion of
// method BlbmeFile on bn instbnce of MockClient.
type ClientBlbmeFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 *BlbmeOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*Hunk
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientBlbmeFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientBlbmeFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientBrbnchesContbiningFunc describes the behbvior when the
// BrbnchesContbining method of the pbrent MockClient instbnce is invoked.
type ClientBrbnchesContbiningFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)
	history     []ClientBrbnchesContbiningFuncCbll
	mutex       sync.Mutex
}

// BrbnchesContbining delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) BrbnchesContbining(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) ([]string, error) {
	r0, r1 := m.BrbnchesContbiningFunc.nextHook()(v0, v1, v2, v3)
	m.BrbnchesContbiningFunc.bppendCbll(ClientBrbnchesContbiningFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the BrbnchesContbining
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientBrbnchesContbiningFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BrbnchesContbining method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientBrbnchesContbiningFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientBrbnchesContbiningFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientBrbnchesContbiningFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
		return r0, r1
	})
}

func (f *ClientBrbnchesContbiningFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientBrbnchesContbiningFunc) bppendCbll(r0 ClientBrbnchesContbiningFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientBrbnchesContbiningFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientBrbnchesContbiningFunc) History() []ClientBrbnchesContbiningFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientBrbnchesContbiningFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientBrbnchesContbiningFuncCbll is bn object thbt describes bn
// invocbtion of method BrbnchesContbining on bn instbnce of MockClient.
type ClientBrbnchesContbiningFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientBrbnchesContbiningFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientBrbnchesContbiningFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitDbteFunc describes the behbvior when the CommitDbte method of
// the pbrent MockClient instbnce is invoked.
type ClientCommitDbteFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)
	history     []ClientCommitDbteFuncCbll
	mutex       sync.Mutex
}

// CommitDbte delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitDbte(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) (string, time.Time, bool, error) {
	r0, r1, r2, r3 := m.CommitDbteFunc.nextHook()(v0, v1, v2, v3)
	m.CommitDbteFunc.bppendCbll(ClientCommitDbteFuncCbll{v0, v1, v2, v3, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the CommitDbte method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitDbteFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitDbte method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientCommitDbteFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitDbteFunc) SetDefbultReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitDbteFunc) PushReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *ClientCommitDbteFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (string, time.Time, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitDbteFunc) bppendCbll(r0 ClientCommitDbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitDbteFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientCommitDbteFunc) History() []ClientCommitDbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitDbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitDbteFuncCbll is bn object thbt describes bn invocbtion of
// method CommitDbte on bn instbnce of MockClient.
type ClientCommitDbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 time.Time
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitDbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitDbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// ClientCommitExistsFunc describes the behbvior when the CommitExists
// method of the pbrent MockClient instbnce is invoked.
type ClientCommitExistsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)
	history     []ClientCommitExistsFuncCbll
	mutex       sync.Mutex
}

// CommitExists delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitExists(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID) (bool, error) {
	r0, r1 := m.CommitExistsFunc.nextHook()(v0, v1, v2, v3)
	m.CommitExistsFunc.bppendCbll(ClientCommitExistsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitExists method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitExistsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitExists method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientCommitExistsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitExistsFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitExistsFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
		return r0, r1
	})
}

func (f *ClientCommitExistsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitExistsFunc) bppendCbll(r0 ClientCommitExistsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitExistsFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientCommitExistsFunc) History() []ClientCommitExistsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitExistsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitExistsFuncCbll is bn object thbt describes bn invocbtion of
// method CommitExists on bn instbnce of MockClient.
type ClientCommitExistsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitExistsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitExistsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitGrbphFunc describes the behbvior when the CommitGrbph method
// of the pbrent MockClient instbnce is invoked.
type ClientCommitGrbphFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error)
	hooks       []func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error)
	history     []ClientCommitGrbphFuncCbll
	mutex       sync.Mutex
}

// CommitGrbph delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitGrbph(v0 context.Context, v1 bpi.RepoNbme, v2 CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
	r0, r1 := m.CommitGrbphFunc.nextHook()(v0, v1, v2)
	m.CommitGrbphFunc.bppendCbll(ClientCommitGrbphFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitGrbph method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitGrbphFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitGrbph method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientCommitGrbphFunc) PushHook(hook func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitGrbphFunc) SetDefbultReturn(r0 *gitdombin.CommitGrbph, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitGrbphFunc) PushReturn(r0 *gitdombin.CommitGrbph, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
		return r0, r1
	})
}

func (f *ClientCommitGrbphFunc) nextHook() func(context.Context, bpi.RepoNbme, CommitGrbphOptions) (*gitdombin.CommitGrbph, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitGrbphFunc) bppendCbll(r0 ClientCommitGrbphFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitGrbphFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientCommitGrbphFunc) History() []ClientCommitGrbphFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitGrbphFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitGrbphFuncCbll is bn object thbt describes bn invocbtion of
// method CommitGrbph on bn instbnce of MockClient.
type ClientCommitGrbphFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 CommitGrbphOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.CommitGrbph
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitGrbphFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitGrbphFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitLogFunc describes the behbvior when the CommitLog method of
// the pbrent MockClient instbnce is invoked.
type ClientCommitLogFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error)
	hooks       []func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error)
	history     []ClientCommitLogFuncCbll
	mutex       sync.Mutex
}

// CommitLog delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitLog(v0 context.Context, v1 bpi.RepoNbme, v2 time.Time) ([]CommitLog, error) {
	r0, r1 := m.CommitLogFunc.nextHook()(v0, v1, v2)
	m.CommitLogFunc.bppendCbll(ClientCommitLogFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitLog method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitLogFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitLog method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientCommitLogFunc) PushHook(hook func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitLogFunc) SetDefbultReturn(r0 []CommitLog, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitLogFunc) PushReturn(r0 []CommitLog, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error) {
		return r0, r1
	})
}

func (f *ClientCommitLogFunc) nextHook() func(context.Context, bpi.RepoNbme, time.Time) ([]CommitLog, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitLogFunc) bppendCbll(r0 ClientCommitLogFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitLogFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientCommitLogFunc) History() []ClientCommitLogFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitLogFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitLogFuncCbll is bn object thbt describes bn invocbtion of
// method CommitLog on bn instbnce of MockClient.
type ClientCommitLogFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []CommitLog
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitLogFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitLogFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitsFunc describes the behbvior when the Commits method of the
// pbrent MockClient instbnce is invoked.
type ClientCommitsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error)
	history     []ClientCommitsFuncCbll
	mutex       sync.Mutex
}

// Commits delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Commits(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 CommitsOptions) ([]*gitdombin.Commit, error) {
	r0, r1 := m.CommitsFunc.nextHook()(v0, v1, v2, v3)
	m.CommitsFunc.bppendCbll(ClientCommitsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Commits method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Commits method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientCommitsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitsFunc) SetDefbultReturn(r0 []*gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitsFunc) PushReturn(r0 []*gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *ClientCommitsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, CommitsOptions) ([]*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitsFunc) bppendCbll(r0 ClientCommitsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitsFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientCommitsFunc) History() []ClientCommitsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitsFuncCbll is bn object thbt describes bn invocbtion of method
// Commits on bn instbnce of MockClient.
type ClientCommitsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 CommitsOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitsExistFunc describes the behbvior when the CommitsExist
// method of the pbrent MockClient instbnce is invoked.
type ClientCommitsExistFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)
	history     []ClientCommitsExistFuncCbll
	mutex       sync.Mutex
}

// CommitsExist delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitsExist(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 []bpi.RepoCommit) ([]bool, error) {
	r0, r1 := m.CommitsExistFunc.nextHook()(v0, v1, v2)
	m.CommitsExistFunc.bppendCbll(ClientCommitsExistFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CommitsExist method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientCommitsExistFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitsExist method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientCommitsExistFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitsExistFunc) SetDefbultReturn(r0 []bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitsExistFunc) PushReturn(r0 []bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
		return r0, r1
	})
}

func (f *ClientCommitsExistFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit) ([]bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitsExistFunc) bppendCbll(r0 ClientCommitsExistFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitsExistFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientCommitsExistFunc) History() []ClientCommitsExistFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitsExistFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitsExistFuncCbll is bn object thbt describes bn invocbtion of
// method CommitsExist on bn instbnce of MockClient.
type ClientCommitsExistFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []bpi.RepoCommit
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitsExistFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitsExistFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCommitsUniqueToBrbnchFunc describes the behbvior when the
// CommitsUniqueToBrbnch method of the pbrent MockClient instbnce is
// invoked.
type ClientCommitsUniqueToBrbnchFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)
	history     []ClientCommitsUniqueToBrbnchFuncCbll
	mutex       sync.Mutex
}

// CommitsUniqueToBrbnch delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CommitsUniqueToBrbnch(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 bool, v5 *time.Time) (mbp[string]time.Time, error) {
	r0, r1 := m.CommitsUniqueToBrbnchFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.CommitsUniqueToBrbnchFunc.bppendCbll(ClientCommitsUniqueToBrbnchFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CommitsUniqueToBrbnch method of the pbrent MockClient instbnce is invoked
// bnd the hook queue is empty.
func (f *ClientCommitsUniqueToBrbnchFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CommitsUniqueToBrbnch method of the pbrent MockClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ClientCommitsUniqueToBrbnchFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCommitsUniqueToBrbnchFunc) SetDefbultReturn(r0 mbp[string]time.Time, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCommitsUniqueToBrbnchFunc) PushReturn(r0 mbp[string]time.Time, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
		return r0, r1
	})
}

func (f *ClientCommitsUniqueToBrbnchFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, bool, *time.Time) (mbp[string]time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitsUniqueToBrbnchFunc) bppendCbll(r0 ClientCommitsUniqueToBrbnchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCommitsUniqueToBrbnchFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientCommitsUniqueToBrbnchFunc) History() []ClientCommitsUniqueToBrbnchFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCommitsUniqueToBrbnchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitsUniqueToBrbnchFuncCbll is bn object thbt describes bn
// invocbtion of method CommitsUniqueToBrbnch on bn instbnce of MockClient.
type ClientCommitsUniqueToBrbnchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 bool
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 *time.Time
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]time.Time
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCommitsUniqueToBrbnchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCommitsUniqueToBrbnchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientContributorCountFunc describes the behbvior when the
// ContributorCount method of the pbrent MockClient instbnce is invoked.
type ClientContributorCountFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error)
	hooks       []func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error)
	history     []ClientContributorCountFuncCbll
	mutex       sync.Mutex
}

// ContributorCount delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ContributorCount(v0 context.Context, v1 bpi.RepoNbme, v2 ContributorOptions) ([]*gitdombin.ContributorCount, error) {
	r0, r1 := m.ContributorCountFunc.nextHook()(v0, v1, v2)
	m.ContributorCountFunc.bppendCbll(ClientContributorCountFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ContributorCount
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientContributorCountFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ContributorCount method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientContributorCountFunc) PushHook(hook func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientContributorCountFunc) SetDefbultReturn(r0 []*gitdombin.ContributorCount, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientContributorCountFunc) PushReturn(r0 []*gitdombin.ContributorCount, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error) {
		return r0, r1
	})
}

func (f *ClientContributorCountFunc) nextHook() func(context.Context, bpi.RepoNbme, ContributorOptions) ([]*gitdombin.ContributorCount, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientContributorCountFunc) bppendCbll(r0 ClientContributorCountFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientContributorCountFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientContributorCountFunc) History() []ClientContributorCountFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientContributorCountFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientContributorCountFuncCbll is bn object thbt describes bn invocbtion
// of method ContributorCount on bn instbnce of MockClient.
type ClientContributorCountFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 ContributorOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.ContributorCount
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientContributorCountFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientContributorCountFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientCrebteCommitFromPbtchFunc describes the behbvior when the
// CrebteCommitFromPbtch method of the pbrent MockClient instbnce is
// invoked.
type ClientCrebteCommitFromPbtchFunc struct {
	defbultHook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)
	hooks       []func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)
	history     []ClientCrebteCommitFromPbtchFuncCbll
	mutex       sync.Mutex
}

// CrebteCommitFromPbtch delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) CrebteCommitFromPbtch(v0 context.Context, v1 protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	r0, r1 := m.CrebteCommitFromPbtchFunc.nextHook()(v0, v1)
	m.CrebteCommitFromPbtchFunc.bppendCbll(ClientCrebteCommitFromPbtchFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteCommitFromPbtch method of the pbrent MockClient instbnce is invoked
// bnd the hook queue is empty.
func (f *ClientCrebteCommitFromPbtchFunc) SetDefbultHook(hook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteCommitFromPbtch method of the pbrent MockClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ClientCrebteCommitFromPbtchFunc) PushHook(hook func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientCrebteCommitFromPbtchFunc) SetDefbultReturn(r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientCrebteCommitFromPbtchFunc) PushReturn(r0 *protocol.CrebteCommitFromPbtchResponse, r1 error) {
	f.PushHook(func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
		return r0, r1
	})
}

func (f *ClientCrebteCommitFromPbtchFunc) nextHook() func(context.Context, protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCrebteCommitFromPbtchFunc) bppendCbll(r0 ClientCrebteCommitFromPbtchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientCrebteCommitFromPbtchFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientCrebteCommitFromPbtchFunc) History() []ClientCrebteCommitFromPbtchFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientCrebteCommitFromPbtchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCrebteCommitFromPbtchFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteCommitFromPbtch on bn instbnce of MockClient.
type ClientCrebteCommitFromPbtchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 protocol.CrebteCommitFromPbtchRequest
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.CrebteCommitFromPbtchResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientCrebteCommitFromPbtchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientCrebteCommitFromPbtchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientDiffFunc describes the behbvior when the Diff method of the pbrent
// MockClient instbnce is invoked.
type ClientDiffFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error)
	history     []ClientDiffFuncCbll
	mutex       sync.Mutex
}

// Diff delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Diff(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 DiffOptions) (*DiffFileIterbtor, error) {
	r0, r1 := m.DiffFunc.nextHook()(v0, v1, v2)
	m.DiffFunc.bppendCbll(ClientDiffFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Diff method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientDiffFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Diff method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientDiffFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientDiffFunc) SetDefbultReturn(r0 *DiffFileIterbtor, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientDiffFunc) PushReturn(r0 *DiffFileIterbtor, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error) {
		return r0, r1
	})
}

func (f *ClientDiffFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, DiffOptions) (*DiffFileIterbtor, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientDiffFunc) bppendCbll(r0 ClientDiffFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientDiffFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientDiffFunc) History() []ClientDiffFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientDiffFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientDiffFuncCbll is bn object thbt describes bn invocbtion of method
// Diff on bn instbnce of MockClient.
type ClientDiffFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 DiffOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *DiffFileIterbtor
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientDiffFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientDiffFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientDiffPbthFunc describes the behbvior when the DiffPbth method of the
// pbrent MockClient instbnce is invoked.
type ClientDiffPbthFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)
	history     []ClientDiffPbthFuncCbll
	mutex       sync.Mutex
}

// DiffPbth delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) DiffPbth(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 string, v5 string) ([]*diff.Hunk, error) {
	r0, r1 := m.DiffPbthFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.DiffPbthFunc.bppendCbll(ClientDiffPbthFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DiffPbth method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientDiffPbthFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DiffPbth method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientDiffPbthFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientDiffPbthFunc) SetDefbultReturn(r0 []*diff.Hunk, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientDiffPbthFunc) PushReturn(r0 []*diff.Hunk, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
		return r0, r1
	})
}

func (f *ClientDiffPbthFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string, string) ([]*diff.Hunk, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientDiffPbthFunc) bppendCbll(r0 ClientDiffPbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientDiffPbthFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientDiffPbthFunc) History() []ClientDiffPbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientDiffPbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientDiffPbthFuncCbll is bn object thbt describes bn invocbtion of
// method DiffPbth on bn instbnce of MockClient.
type ClientDiffPbthFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*diff.Hunk
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientDiffPbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientDiffPbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientDiffSymbolsFunc describes the behbvior when the DiffSymbols method
// of the pbrent MockClient instbnce is invoked.
type ClientDiffSymbolsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)
	history     []ClientDiffSymbolsFuncCbll
	mutex       sync.Mutex
}

// DiffSymbols delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) DiffSymbols(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 bpi.CommitID) ([]byte, error) {
	r0, r1 := m.DiffSymbolsFunc.nextHook()(v0, v1, v2, v3)
	m.DiffSymbolsFunc.bppendCbll(ClientDiffSymbolsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DiffSymbols method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientDiffSymbolsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DiffSymbols method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientDiffSymbolsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientDiffSymbolsFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientDiffSymbolsFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
		return r0, r1
	})
}

func (f *ClientDiffSymbolsFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientDiffSymbolsFunc) bppendCbll(r0 ClientDiffSymbolsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientDiffSymbolsFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientDiffSymbolsFunc) History() []ClientDiffSymbolsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientDiffSymbolsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientDiffSymbolsFuncCbll is bn object thbt describes bn invocbtion of
// method DiffSymbols on bn instbnce of MockClient.
type ClientDiffSymbolsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientDiffSymbolsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientDiffSymbolsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientFirstEverCommitFunc describes the behbvior when the FirstEverCommit
// method of the pbrent MockClient instbnce is invoked.
type ClientFirstEverCommitFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)
	history     []ClientFirstEverCommitFuncCbll
	mutex       sync.Mutex
}

// FirstEverCommit delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) FirstEverCommit(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme) (*gitdombin.Commit, error) {
	r0, r1 := m.FirstEverCommitFunc.nextHook()(v0, v1, v2)
	m.FirstEverCommitFunc.bppendCbll(ClientFirstEverCommitFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the FirstEverCommit
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientFirstEverCommitFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// FirstEverCommit method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientFirstEverCommitFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientFirstEverCommitFunc) SetDefbultReturn(r0 *gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientFirstEverCommitFunc) PushReturn(r0 *gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *ClientFirstEverCommitFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientFirstEverCommitFunc) bppendCbll(r0 ClientFirstEverCommitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientFirstEverCommitFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientFirstEverCommitFunc) History() []ClientFirstEverCommitFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientFirstEverCommitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientFirstEverCommitFuncCbll is bn object thbt describes bn invocbtion
// of method FirstEverCommit on bn instbnce of MockClient.
type ClientFirstEverCommitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientFirstEverCommitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientFirstEverCommitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetBehindAhebdFunc describes the behbvior when the GetBehindAhebd
// method of the pbrent MockClient instbnce is invoked.
type ClientGetBehindAhebdFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)
	history     []ClientGetBehindAhebdFuncCbll
	mutex       sync.Mutex
}

// GetBehindAhebd delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetBehindAhebd(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 string) (*gitdombin.BehindAhebd, error) {
	r0, r1 := m.GetBehindAhebdFunc.nextHook()(v0, v1, v2, v3)
	m.GetBehindAhebdFunc.bppendCbll(ClientGetBehindAhebdFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBehindAhebd
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientGetBehindAhebdFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBehindAhebd method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetBehindAhebdFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetBehindAhebdFunc) SetDefbultReturn(r0 *gitdombin.BehindAhebd, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetBehindAhebdFunc) PushReturn(r0 *gitdombin.BehindAhebd, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
		return r0, r1
	})
}

func (f *ClientGetBehindAhebdFunc) nextHook() func(context.Context, bpi.RepoNbme, string, string) (*gitdombin.BehindAhebd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetBehindAhebdFunc) bppendCbll(r0 ClientGetBehindAhebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetBehindAhebdFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientGetBehindAhebdFunc) History() []ClientGetBehindAhebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetBehindAhebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetBehindAhebdFuncCbll is bn object thbt describes bn invocbtion of
// method GetBehindAhebd on bn instbnce of MockClient.
type ClientGetBehindAhebdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.BehindAhebd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetBehindAhebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetBehindAhebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetCommitFunc describes the behbvior when the GetCommit method of
// the pbrent MockClient instbnce is invoked.
type ClientGetCommitFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error)
	history     []ClientGetCommitFuncCbll
	mutex       sync.Mutex
}

// GetCommit delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetCommit(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 ResolveRevisionOptions) (*gitdombin.Commit, error) {
	r0, r1 := m.GetCommitFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetCommitFunc.bppendCbll(ClientGetCommitFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetCommit method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientGetCommitFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommit method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetCommitFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetCommitFunc) SetDefbultReturn(r0 *gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetCommitFunc) PushReturn(r0 *gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *ClientGetCommitFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ResolveRevisionOptions) (*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetCommitFunc) bppendCbll(r0 ClientGetCommitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetCommitFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientGetCommitFunc) History() []ClientGetCommitFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetCommitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetCommitFuncCbll is bn object thbt describes bn invocbtion of
// method GetCommit on bn instbnce of MockClient.
type ClientGetCommitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 ResolveRevisionOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetCommitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetCommitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetCommitsFunc describes the behbvior when the GetCommits method of
// the pbrent MockClient instbnce is invoked.
type ClientGetCommitsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)
	history     []ClientGetCommitsFuncCbll
	mutex       sync.Mutex
}

// GetCommits delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetCommits(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 []bpi.RepoCommit, v3 bool) ([]*gitdombin.Commit, error) {
	r0, r1 := m.GetCommitsFunc.nextHook()(v0, v1, v2, v3)
	m.GetCommitsFunc.bppendCbll(ClientGetCommitsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetCommits method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientGetCommitsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommits method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetCommitsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetCommitsFunc) SetDefbultReturn(r0 []*gitdombin.Commit, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetCommitsFunc) PushReturn(r0 []*gitdombin.Commit, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
		return r0, r1
	})
}

func (f *ClientGetCommitsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, []bpi.RepoCommit, bool) ([]*gitdombin.Commit, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetCommitsFunc) bppendCbll(r0 ClientGetCommitsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetCommitsFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientGetCommitsFunc) History() []ClientGetCommitsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetCommitsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetCommitsFuncCbll is bn object thbt describes bn invocbtion of
// method GetCommits on bn instbnce of MockClient.
type ClientGetCommitsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []bpi.RepoCommit
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Commit
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetCommitsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetCommitsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientGetDefbultBrbnchFunc describes the behbvior when the
// GetDefbultBrbnch method of the pbrent MockClient instbnce is invoked.
type ClientGetDefbultBrbnchFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)
	history     []ClientGetDefbultBrbnchFuncCbll
	mutex       sync.Mutex
}

// GetDefbultBrbnch delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetDefbultBrbnch(v0 context.Context, v1 bpi.RepoNbme, v2 bool) (string, bpi.CommitID, error) {
	r0, r1, r2 := m.GetDefbultBrbnchFunc.nextHook()(v0, v1, v2)
	m.GetDefbultBrbnchFunc.bppendCbll(ClientGetDefbultBrbnchFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetDefbultBrbnch
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientGetDefbultBrbnchFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDefbultBrbnch method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientGetDefbultBrbnchFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetDefbultBrbnchFunc) SetDefbultReturn(r0 string, r1 bpi.CommitID, r2 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetDefbultBrbnchFunc) PushReturn(r0 string, r1 bpi.CommitID, r2 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
		return r0, r1, r2
	})
}

func (f *ClientGetDefbultBrbnchFunc) nextHook() func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetDefbultBrbnchFunc) bppendCbll(r0 ClientGetDefbultBrbnchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetDefbultBrbnchFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientGetDefbultBrbnchFunc) History() []ClientGetDefbultBrbnchFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetDefbultBrbnchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetDefbultBrbnchFuncCbll is bn object thbt describes bn invocbtion
// of method GetDefbultBrbnch on bn instbnce of MockClient.
type ClientGetDefbultBrbnchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bpi.CommitID
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetDefbultBrbnchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetDefbultBrbnchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientGetObjectFunc describes the behbvior when the GetObject method of
// the pbrent MockClient instbnce is invoked.
type ClientGetObjectFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)
	hooks       []func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)
	history     []ClientGetObjectFuncCbll
	mutex       sync.Mutex
}

// GetObject delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) GetObject(v0 context.Context, v1 bpi.RepoNbme, v2 string) (*gitdombin.GitObject, error) {
	r0, r1 := m.GetObjectFunc.nextHook()(v0, v1, v2)
	m.GetObjectFunc.bppendCbll(ClientGetObjectFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetObject method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientGetObjectFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetObject method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientGetObjectFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientGetObjectFunc) SetDefbultReturn(r0 *gitdombin.GitObject, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientGetObjectFunc) PushReturn(r0 *gitdombin.GitObject, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
		return r0, r1
	})
}

func (f *ClientGetObjectFunc) nextHook() func(context.Context, bpi.RepoNbme, string) (*gitdombin.GitObject, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientGetObjectFunc) bppendCbll(r0 ClientGetObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientGetObjectFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientGetObjectFunc) History() []ClientGetObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientGetObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientGetObjectFuncCbll is bn object thbt describes bn invocbtion of
// method GetObject on bn instbnce of MockClient.
type ClientGetObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *gitdombin.GitObject
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientGetObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientGetObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientHbsCommitAfterFunc describes the behbvior when the HbsCommitAfter
// method of the pbrent MockClient instbnce is invoked.
type ClientHbsCommitAfterFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)
	history     []ClientHbsCommitAfterFuncCbll
	mutex       sync.Mutex
}

// HbsCommitAfter delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) HbsCommitAfter(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 string) (bool, error) {
	r0, r1 := m.HbsCommitAfterFunc.nextHook()(v0, v1, v2, v3, v4)
	m.HbsCommitAfterFunc.bppendCbll(ClientHbsCommitAfterFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the HbsCommitAfter
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientHbsCommitAfterFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HbsCommitAfter method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientHbsCommitAfterFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientHbsCommitAfterFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientHbsCommitAfterFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
		return r0, r1
	})
}

func (f *ClientHbsCommitAfterFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientHbsCommitAfterFunc) bppendCbll(r0 ClientHbsCommitAfterFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientHbsCommitAfterFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientHbsCommitAfterFunc) History() []ClientHbsCommitAfterFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientHbsCommitAfterFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientHbsCommitAfterFuncCbll is bn object thbt describes bn invocbtion of
// method HbsCommitAfter on bn instbnce of MockClient.
type ClientHbsCommitAfterFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientHbsCommitAfterFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientHbsCommitAfterFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientHebdFunc describes the behbvior when the Hebd method of the pbrent
// MockClient instbnce is invoked.
type ClientHebdFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)
	history     []ClientHebdFuncCbll
	mutex       sync.Mutex
}

// Hebd delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Hebd(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme) (string, bool, error) {
	r0, r1, r2 := m.HebdFunc.nextHook()(v0, v1, v2)
	m.HebdFunc.bppendCbll(ClientHebdFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the Hebd method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientHebdFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hebd method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientHebdFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientHebdFunc) SetDefbultReturn(r0 string, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientHebdFunc) PushReturn(r0 string, r1 bool, r2 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientHebdFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme) (string, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientHebdFunc) bppendCbll(r0 ClientHebdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientHebdFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientHebdFunc) History() []ClientHebdFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientHebdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientHebdFuncCbll is bn object thbt describes bn invocbtion of method
// Hebd on bn instbnce of MockClient.
type ClientHebdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientHebdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientHebdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientIsRepoClonebbleFunc describes the behbvior when the IsRepoClonebble
// method of the pbrent MockClient instbnce is invoked.
type ClientIsRepoClonebbleFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) error
	hooks       []func(context.Context, bpi.RepoNbme) error
	history     []ClientIsRepoClonebbleFuncCbll
	mutex       sync.Mutex
}

// IsRepoClonebble delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) IsRepoClonebble(v0 context.Context, v1 bpi.RepoNbme) error {
	r0 := m.IsRepoClonebbleFunc.nextHook()(v0, v1)
	m.IsRepoClonebbleFunc.bppendCbll(ClientIsRepoClonebbleFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the IsRepoClonebble
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientIsRepoClonebbleFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IsRepoClonebble method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientIsRepoClonebbleFunc) PushHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientIsRepoClonebbleFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientIsRepoClonebbleFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

func (f *ClientIsRepoClonebbleFunc) nextHook() func(context.Context, bpi.RepoNbme) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientIsRepoClonebbleFunc) bppendCbll(r0 ClientIsRepoClonebbleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientIsRepoClonebbleFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientIsRepoClonebbleFunc) History() []ClientIsRepoClonebbleFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientIsRepoClonebbleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientIsRepoClonebbleFuncCbll is bn object thbt describes bn invocbtion
// of method IsRepoClonebble on bn instbnce of MockClient.
type ClientIsRepoClonebbleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientIsRepoClonebbleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientIsRepoClonebbleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientListBrbnchesFunc describes the behbvior when the ListBrbnches
// method of the pbrent MockClient instbnce is invoked.
type ClientListBrbnchesFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error)
	hooks       []func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error)
	history     []ClientListBrbnchesFuncCbll
	mutex       sync.Mutex
}

// ListBrbnches delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListBrbnches(v0 context.Context, v1 bpi.RepoNbme, v2 BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
	r0, r1 := m.ListBrbnchesFunc.nextHook()(v0, v1, v2)
	m.ListBrbnchesFunc.bppendCbll(ClientListBrbnchesFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListBrbnches method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientListBrbnchesFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListBrbnches method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientListBrbnchesFunc) PushHook(hook func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListBrbnchesFunc) SetDefbultReturn(r0 []*gitdombin.Brbnch, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListBrbnchesFunc) PushReturn(r0 []*gitdombin.Brbnch, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
		return r0, r1
	})
}

func (f *ClientListBrbnchesFunc) nextHook() func(context.Context, bpi.RepoNbme, BrbnchesOptions) ([]*gitdombin.Brbnch, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListBrbnchesFunc) bppendCbll(r0 ClientListBrbnchesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListBrbnchesFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListBrbnchesFunc) History() []ClientListBrbnchesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListBrbnchesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListBrbnchesFuncCbll is bn object thbt describes bn invocbtion of
// method ListBrbnches on bn instbnce of MockClient.
type ClientListBrbnchesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 BrbnchesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Brbnch
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListBrbnchesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListBrbnchesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientListDirectoryChildrenFunc describes the behbvior when the
// ListDirectoryChildren method of the pbrent MockClient instbnce is
// invoked.
type ClientListDirectoryChildrenFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)
	history     []ClientListDirectoryChildrenFuncCbll
	mutex       sync.Mutex
}

// ListDirectoryChildren delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListDirectoryChildren(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 []string) (mbp[string][]string, error) {
	r0, r1 := m.ListDirectoryChildrenFunc.nextHook()(v0, v1, v2, v3, v4)
	m.ListDirectoryChildrenFunc.bppendCbll(ClientListDirectoryChildrenFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListDirectoryChildren method of the pbrent MockClient instbnce is invoked
// bnd the hook queue is empty.
func (f *ClientListDirectoryChildrenFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListDirectoryChildren method of the pbrent MockClient instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *ClientListDirectoryChildrenFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListDirectoryChildrenFunc) SetDefbultReturn(r0 mbp[string][]string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListDirectoryChildrenFunc) PushReturn(r0 mbp[string][]string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
		return r0, r1
	})
}

func (f *ClientListDirectoryChildrenFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, []string) (mbp[string][]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListDirectoryChildrenFunc) bppendCbll(r0 ClientListDirectoryChildrenFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListDirectoryChildrenFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientListDirectoryChildrenFunc) History() []ClientListDirectoryChildrenFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListDirectoryChildrenFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListDirectoryChildrenFuncCbll is bn object thbt describes bn
// invocbtion of method ListDirectoryChildren on bn instbnce of MockClient.
type ClientListDirectoryChildrenFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string][]string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListDirectoryChildrenFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListDirectoryChildrenFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientListRefsFunc describes the behbvior when the ListRefs method of the
// pbrent MockClient instbnce is invoked.
type ClientListRefsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)
	hooks       []func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)
	history     []ClientListRefsFuncCbll
	mutex       sync.Mutex
}

// ListRefs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListRefs(v0 context.Context, v1 bpi.RepoNbme) ([]gitdombin.Ref, error) {
	r0, r1 := m.ListRefsFunc.nextHook()(v0, v1)
	m.ListRefsFunc.bppendCbll(ClientListRefsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListRefs method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientListRefsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRefs method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientListRefsFunc) PushHook(hook func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListRefsFunc) SetDefbultReturn(r0 []gitdombin.Ref, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListRefsFunc) PushReturn(r0 []gitdombin.Ref, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
		return r0, r1
	})
}

func (f *ClientListRefsFunc) nextHook() func(context.Context, bpi.RepoNbme) ([]gitdombin.Ref, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListRefsFunc) bppendCbll(r0 ClientListRefsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListRefsFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientListRefsFunc) History() []ClientListRefsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListRefsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListRefsFuncCbll is bn object thbt describes bn invocbtion of
// method ListRefs on bn instbnce of MockClient.
type ClientListRefsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []gitdombin.Ref
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientListRefsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListRefsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientListTbgsFunc describes the behbvior when the ListTbgs method of the
// pbrent MockClient instbnce is invoked.
type ClientListTbgsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)
	hooks       []func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)
	history     []ClientListTbgsFuncCbll
	mutex       sync.Mutex
}

// ListTbgs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ListTbgs(v0 context.Context, v1 bpi.RepoNbme, v2 ...string) ([]*gitdombin.Tbg, error) {
	r0, r1 := m.ListTbgsFunc.nextHook()(v0, v1, v2...)
	m.ListTbgsFunc.bppendCbll(ClientListTbgsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListTbgs method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientListTbgsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListTbgs method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientListTbgsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientListTbgsFunc) SetDefbultReturn(r0 []*gitdombin.Tbg, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientListTbgsFunc) PushReturn(r0 []*gitdombin.Tbg, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
		return r0, r1
	})
}

func (f *ClientListTbgsFunc) nextHook() func(context.Context, bpi.RepoNbme, ...string) ([]*gitdombin.Tbg, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientListTbgsFunc) bppendCbll(r0 ClientListTbgsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientListTbgsFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientListTbgsFunc) History() []ClientListTbgsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientListTbgsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientListTbgsFuncCbll is bn object thbt describes bn invocbtion of
// method ListTbgs on bn instbnce of MockClient.
type ClientListTbgsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*gitdombin.Tbg
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientListTbgsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientListTbgsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientLogReverseEbchFunc describes the behbvior when the LogReverseEbch
// method of the pbrent MockClient instbnce is invoked.
type ClientLogReverseEbchFunc struct {
	defbultHook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error
	hooks       []func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error
	history     []ClientLogReverseEbchFuncCbll
	mutex       sync.Mutex
}

// LogReverseEbch delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) LogReverseEbch(v0 context.Context, v1 string, v2 string, v3 int, v4 func(entry gitdombin.LogEntry) error) error {
	r0 := m.LogReverseEbchFunc.nextHook()(v0, v1, v2, v3, v4)
	m.LogReverseEbchFunc.bppendCbll(ClientLogReverseEbchFuncCbll{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LogReverseEbch
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientLogReverseEbchFunc) SetDefbultHook(hook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LogReverseEbch method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientLogReverseEbchFunc) PushHook(hook func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientLogReverseEbchFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientLogReverseEbchFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
		return r0
	})
}

func (f *ClientLogReverseEbchFunc) nextHook() func(context.Context, string, string, int, func(entry gitdombin.LogEntry) error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientLogReverseEbchFunc) bppendCbll(r0 ClientLogReverseEbchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientLogReverseEbchFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientLogReverseEbchFunc) History() []ClientLogReverseEbchFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientLogReverseEbchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientLogReverseEbchFuncCbll is bn object thbt describes bn invocbtion of
// method LogReverseEbch on bn instbnce of MockClient.
type ClientLogReverseEbchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 func(entry gitdombin.LogEntry) error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientLogReverseEbchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientLogReverseEbchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientLsFilesFunc describes the behbvior when the LsFiles method of the
// pbrent MockClient instbnce is invoked.
type ClientLsFilesFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)
	history     []ClientLsFilesFuncCbll
	mutex       sync.Mutex
}

// LsFiles delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) LsFiles(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 ...gitdombin.Pbthspec) ([]string, error) {
	r0, r1 := m.LsFilesFunc.nextHook()(v0, v1, v2, v3, v4...)
	m.LsFilesFunc.bppendCbll(ClientLsFilesFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LsFiles method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientLsFilesFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LsFiles method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientLsFilesFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientLsFilesFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientLsFilesFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
		return r0, r1
	})
}

func (f *ClientLsFilesFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, ...gitdombin.Pbthspec) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientLsFilesFunc) bppendCbll(r0 ClientLsFilesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientLsFilesFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientLsFilesFunc) History() []ClientLsFilesFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientLsFilesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientLsFilesFuncCbll is bn object thbt describes bn invocbtion of method
// LsFiles on bn instbnce of MockClient.
type ClientLsFilesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg4 []gitdombin.Pbthspec
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientLsFilesFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg4 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientLsFilesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientMergeBbseFunc describes the behbvior when the MergeBbse method of
// the pbrent MockClient instbnce is invoked.
type ClientMergeBbseFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)
	history     []ClientMergeBbseFuncCbll
	mutex       sync.Mutex
}

// MergeBbse delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) MergeBbse(v0 context.Context, v1 bpi.RepoNbme, v2 bpi.CommitID, v3 bpi.CommitID) (bpi.CommitID, error) {
	r0, r1 := m.MergeBbseFunc.nextHook()(v0, v1, v2, v3)
	m.MergeBbseFunc.bppendCbll(ClientMergeBbseFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the MergeBbse method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientMergeBbseFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MergeBbse method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientMergeBbseFunc) PushHook(hook func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientMergeBbseFunc) SetDefbultReturn(r0 bpi.CommitID, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientMergeBbseFunc) PushReturn(r0 bpi.CommitID, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
		return r0, r1
	})
}

func (f *ClientMergeBbseFunc) nextHook() func(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientMergeBbseFunc) bppendCbll(r0 ClientMergeBbseFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientMergeBbseFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientMergeBbseFunc) History() []ClientMergeBbseFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientMergeBbseFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientMergeBbseFuncCbll is bn object thbt describes bn invocbtion of
// method MergeBbse on bn instbnce of MockClient.
type ClientMergeBbseFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bpi.CommitID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientMergeBbseFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientMergeBbseFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientNewFileRebderFunc describes the behbvior when the NewFileRebder
// method of the pbrent MockClient instbnce is invoked.
type ClientNewFileRebderFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)
	history     []ClientNewFileRebderFuncCbll
	mutex       sync.Mutex
}

// NewFileRebder delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) NewFileRebder(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) (io.RebdCloser, error) {
	r0, r1 := m.NewFileRebderFunc.nextHook()(v0, v1, v2, v3, v4)
	m.NewFileRebderFunc.bppendCbll(ClientNewFileRebderFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewFileRebder method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientNewFileRebderFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewFileRebder method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientNewFileRebderFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientNewFileRebderFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientNewFileRebderFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *ClientNewFileRebderFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientNewFileRebderFunc) bppendCbll(r0 ClientNewFileRebderFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientNewFileRebderFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientNewFileRebderFunc) History() []ClientNewFileRebderFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientNewFileRebderFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientNewFileRebderFuncCbll is bn object thbt describes bn invocbtion of
// method NewFileRebder on bn instbnce of MockClient.
type ClientNewFileRebderFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientNewFileRebderFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientNewFileRebderFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientP4ExecFunc describes the behbvior when the P4Exec method of the
// pbrent MockClient instbnce is invoked.
type ClientP4ExecFunc struct {
	defbultHook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)
	hooks       []func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)
	history     []ClientP4ExecFuncCbll
	mutex       sync.Mutex
}

// P4Exec delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) P4Exec(v0 context.Context, v1 string, v2 string, v3 string, v4 ...string) (io.RebdCloser, http.Hebder, error) {
	r0, r1, r2 := m.P4ExecFunc.nextHook()(v0, v1, v2, v3, v4...)
	m.P4ExecFunc.bppendCbll(ClientP4ExecFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the P4Exec method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientP4ExecFunc) SetDefbultHook(hook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// P4Exec method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientP4ExecFunc) PushHook(hook func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientP4ExecFunc) SetDefbultReturn(r0 io.RebdCloser, r1 http.Hebder, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientP4ExecFunc) PushReturn(r0 io.RebdCloser, r1 http.Hebder, r2 error) {
	f.PushHook(func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
		return r0, r1, r2
	})
}

func (f *ClientP4ExecFunc) nextHook() func(context.Context, string, string, string, ...string) (io.RebdCloser, http.Hebder, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientP4ExecFunc) bppendCbll(r0 ClientP4ExecFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientP4ExecFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientP4ExecFunc) History() []ClientP4ExecFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientP4ExecFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientP4ExecFuncCbll is bn object thbt describes bn invocbtion of method
// P4Exec on bn instbnce of MockClient.
type ClientP4ExecFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg4 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 http.Hebder
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientP4ExecFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg4 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientP4ExecFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// ClientP4GetChbngelistFunc describes the behbvior when the P4GetChbngelist
// method of the pbrent MockClient instbnce is invoked.
type ClientP4GetChbngelistFunc struct {
	defbultHook func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error)
	hooks       []func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error)
	history     []ClientP4GetChbngelistFuncCbll
	mutex       sync.Mutex
}

// P4GetChbngelist delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) P4GetChbngelist(v0 context.Context, v1 string, v2 PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
	r0, r1 := m.P4GetChbngelistFunc.nextHook()(v0, v1, v2)
	m.P4GetChbngelistFunc.bppendCbll(ClientP4GetChbngelistFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the P4GetChbngelist
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientP4GetChbngelistFunc) SetDefbultHook(hook func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// P4GetChbngelist method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientP4GetChbngelistFunc) PushHook(hook func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientP4GetChbngelistFunc) SetDefbultReturn(r0 *protocol.PerforceChbngelist, r1 error) {
	f.SetDefbultHook(func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientP4GetChbngelistFunc) PushReturn(r0 *protocol.PerforceChbngelist, r1 error) {
	f.PushHook(func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
		return r0, r1
	})
}

func (f *ClientP4GetChbngelistFunc) nextHook() func(context.Context, string, PerforceCredentibls) (*protocol.PerforceChbngelist, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientP4GetChbngelistFunc) bppendCbll(r0 ClientP4GetChbngelistFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientP4GetChbngelistFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientP4GetChbngelistFunc) History() []ClientP4GetChbngelistFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientP4GetChbngelistFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientP4GetChbngelistFuncCbll is bn object thbt describes bn invocbtion
// of method P4GetChbngelist on bn instbnce of MockClient.
type ClientP4GetChbngelistFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 PerforceCredentibls
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.PerforceChbngelist
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientP4GetChbngelistFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientP4GetChbngelistFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRebdDirFunc describes the behbvior when the RebdDir method of the
// pbrent MockClient instbnce is invoked.
type ClientRebdDirFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)
	history     []ClientRebdDirFuncCbll
	mutex       sync.Mutex
}

// RebdDir delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RebdDir(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string, v5 bool) ([]fs.FileInfo, error) {
	r0, r1 := m.RebdDirFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.RebdDirFunc.bppendCbll(ClientRebdDirFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RebdDir method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientRebdDirFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RebdDir method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientRebdDirFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRebdDirFunc) SetDefbultReturn(r0 []fs.FileInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRebdDirFunc) PushReturn(r0 []fs.FileInfo, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
		return r0, r1
	})
}

func (f *ClientRebdDirFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRebdDirFunc) bppendCbll(r0 ClientRebdDirFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRebdDirFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientRebdDirFunc) History() []ClientRebdDirFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRebdDirFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRebdDirFuncCbll is bn object thbt describes bn invocbtion of method
// RebdDir on bn instbnce of MockClient.
type ClientRebdDirFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []fs.FileInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRebdDirFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRebdDirFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRebdFileFunc describes the behbvior when the RebdFile method of the
// pbrent MockClient instbnce is invoked.
type ClientRebdFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)
	history     []ClientRebdFileFuncCbll
	mutex       sync.Mutex
}

// RebdFile delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RebdFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) ([]byte, error) {
	r0, r1 := m.RebdFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.RebdFileFunc.bppendCbll(ClientRebdFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RebdFile method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientRebdFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RebdFile method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientRebdFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRebdFileFunc) SetDefbultReturn(r0 []byte, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRebdFileFunc) PushReturn(r0 []byte, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
		return r0, r1
	})
}

func (f *ClientRebdFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRebdFileFunc) bppendCbll(r0 ClientRebdFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRebdFileFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientRebdFileFunc) History() []ClientRebdFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRebdFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRebdFileFuncCbll is bn object thbt describes bn invocbtion of
// method RebdFile on bn instbnce of MockClient.
type ClientRebdFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRebdFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRebdFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRefDescriptionsFunc describes the behbvior when the RefDescriptions
// method of the pbrent MockClient instbnce is invoked.
type ClientRefDescriptionsFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)
	history     []ClientRefDescriptionsFuncCbll
	mutex       sync.Mutex
}

// RefDescriptions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RefDescriptions(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 ...string) (mbp[string][]gitdombin.RefDescription, error) {
	r0, r1 := m.RefDescriptionsFunc.nextHook()(v0, v1, v2, v3...)
	m.RefDescriptionsFunc.bppendCbll(ClientRefDescriptionsFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RefDescriptions
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientRefDescriptionsFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RefDescriptions method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientRefDescriptionsFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRefDescriptionsFunc) SetDefbultReturn(r0 mbp[string][]gitdombin.RefDescription, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRefDescriptionsFunc) PushReturn(r0 mbp[string][]gitdombin.RefDescription, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
		return r0, r1
	})
}

func (f *ClientRefDescriptionsFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, ...string) (mbp[string][]gitdombin.RefDescription, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRefDescriptionsFunc) bppendCbll(r0 ClientRefDescriptionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRefDescriptionsFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientRefDescriptionsFunc) History() []ClientRefDescriptionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRefDescriptionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRefDescriptionsFuncCbll is bn object thbt describes bn invocbtion
// of method RefDescriptions on bn instbnce of MockClient.
type ClientRefDescriptionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg3 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string][]gitdombin.RefDescription
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientRefDescriptionsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg3 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1, c.Arg2}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRefDescriptionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRemoveFunc describes the behbvior when the Remove method of the
// pbrent MockClient instbnce is invoked.
type ClientRemoveFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) error
	hooks       []func(context.Context, bpi.RepoNbme) error
	history     []ClientRemoveFuncCbll
	mutex       sync.Mutex
}

// Remove delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Remove(v0 context.Context, v1 bpi.RepoNbme) error {
	r0 := m.RemoveFunc.nextHook()(v0, v1)
	m.RemoveFunc.bppendCbll(ClientRemoveFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Remove method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientRemoveFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Remove method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientRemoveFunc) PushHook(hook func(context.Context, bpi.RepoNbme) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRemoveFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRemoveFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) error {
		return r0
	})
}

func (f *ClientRemoveFunc) nextHook() func(context.Context, bpi.RepoNbme) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRemoveFunc) bppendCbll(r0 ClientRemoveFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRemoveFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientRemoveFunc) History() []ClientRemoveFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRemoveFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRemoveFuncCbll is bn object thbt describes bn invocbtion of method
// Remove on bn instbnce of MockClient.
type ClientRemoveFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRemoveFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRemoveFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientRepoCloneProgressFunc describes the behbvior when the
// RepoCloneProgress method of the pbrent MockClient instbnce is invoked.
type ClientRepoCloneProgressFunc struct {
	defbultHook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)
	hooks       []func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)
	history     []ClientRepoCloneProgressFuncCbll
	mutex       sync.Mutex
}

// RepoCloneProgress delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RepoCloneProgress(v0 context.Context, v1 ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	r0, r1 := m.RepoCloneProgressFunc.nextHook()(v0, v1...)
	m.RepoCloneProgressFunc.bppendCbll(ClientRepoCloneProgressFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoCloneProgress
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientRepoCloneProgressFunc) SetDefbultHook(hook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoCloneProgress method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientRepoCloneProgressFunc) PushHook(hook func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRepoCloneProgressFunc) SetDefbultReturn(r0 *protocol.RepoCloneProgressResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRepoCloneProgressFunc) PushReturn(r0 *protocol.RepoCloneProgressResponse, r1 error) {
	f.PushHook(func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
		return r0, r1
	})
}

func (f *ClientRepoCloneProgressFunc) nextHook() func(context.Context, ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRepoCloneProgressFunc) bppendCbll(r0 ClientRepoCloneProgressFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRepoCloneProgressFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientRepoCloneProgressFunc) History() []ClientRepoCloneProgressFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRepoCloneProgressFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRepoCloneProgressFuncCbll is bn object thbt describes bn invocbtion
// of method RepoCloneProgress on bn instbnce of MockClient.
type ClientRepoCloneProgressFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoCloneProgressResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c ClientRepoCloneProgressFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRepoCloneProgressFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRequestRepoCloneFunc describes the behbvior when the
// RequestRepoClone method of the pbrent MockClient instbnce is invoked.
type ClientRequestRepoCloneFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)
	hooks       []func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)
	history     []ClientRequestRepoCloneFuncCbll
	mutex       sync.Mutex
}

// RequestRepoClone delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RequestRepoClone(v0 context.Context, v1 bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
	r0, r1 := m.RequestRepoCloneFunc.nextHook()(v0, v1)
	m.RequestRepoCloneFunc.bppendCbll(ClientRequestRepoCloneFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RequestRepoClone
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientRequestRepoCloneFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RequestRepoClone method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientRequestRepoCloneFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRequestRepoCloneFunc) SetDefbultReturn(r0 *protocol.RepoCloneResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRequestRepoCloneFunc) PushReturn(r0 *protocol.RepoCloneResponse, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
		return r0, r1
	})
}

func (f *ClientRequestRepoCloneFunc) nextHook() func(context.Context, bpi.RepoNbme) (*protocol.RepoCloneResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRequestRepoCloneFunc) bppendCbll(r0 ClientRequestRepoCloneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRequestRepoCloneFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientRequestRepoCloneFunc) History() []ClientRequestRepoCloneFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRequestRepoCloneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRequestRepoCloneFuncCbll is bn object thbt describes bn invocbtion
// of method RequestRepoClone on bn instbnce of MockClient.
type ClientRequestRepoCloneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoCloneResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRequestRepoCloneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRequestRepoCloneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRequestRepoUpdbteFunc describes the behbvior when the
// RequestRepoUpdbte method of the pbrent MockClient instbnce is invoked.
type ClientRequestRepoUpdbteFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)
	hooks       []func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)
	history     []ClientRequestRepoUpdbteFuncCbll
	mutex       sync.Mutex
}

// RequestRepoUpdbte delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RequestRepoUpdbte(v0 context.Context, v1 bpi.RepoNbme, v2 time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
	r0, r1 := m.RequestRepoUpdbteFunc.nextHook()(v0, v1, v2)
	m.RequestRepoUpdbteFunc.bppendCbll(ClientRequestRepoUpdbteFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RequestRepoUpdbte
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientRequestRepoUpdbteFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RequestRepoUpdbte method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientRequestRepoUpdbteFunc) PushHook(hook func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRequestRepoUpdbteFunc) SetDefbultReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRequestRepoUpdbteFunc) PushReturn(r0 *protocol.RepoUpdbteResponse, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
		return r0, r1
	})
}

func (f *ClientRequestRepoUpdbteFunc) nextHook() func(context.Context, bpi.RepoNbme, time.Durbtion) (*protocol.RepoUpdbteResponse, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRequestRepoUpdbteFunc) bppendCbll(r0 ClientRequestRepoUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRequestRepoUpdbteFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientRequestRepoUpdbteFunc) History() []ClientRequestRepoUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRequestRepoUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRequestRepoUpdbteFuncCbll is bn object thbt describes bn invocbtion
// of method RequestRepoUpdbte on bn instbnce of MockClient.
type ClientRequestRepoUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Durbtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *protocol.RepoUpdbteResponse
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRequestRepoUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRequestRepoUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientResolveRevisionFunc describes the behbvior when the ResolveRevision
// method of the pbrent MockClient instbnce is invoked.
type ClientResolveRevisionFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error)
	hooks       []func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error)
	history     []ClientResolveRevisionFuncCbll
	mutex       sync.Mutex
}

// ResolveRevision delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ResolveRevision(v0 context.Context, v1 bpi.RepoNbme, v2 string, v3 ResolveRevisionOptions) (bpi.CommitID, error) {
	r0, r1 := m.ResolveRevisionFunc.nextHook()(v0, v1, v2, v3)
	m.ResolveRevisionFunc.bppendCbll(ClientResolveRevisionFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveRevision
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientResolveRevisionFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveRevision method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientResolveRevisionFunc) PushHook(hook func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientResolveRevisionFunc) SetDefbultReturn(r0 bpi.CommitID, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientResolveRevisionFunc) PushReturn(r0 bpi.CommitID, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error) {
		return r0, r1
	})
}

func (f *ClientResolveRevisionFunc) nextHook() func(context.Context, bpi.RepoNbme, string, ResolveRevisionOptions) (bpi.CommitID, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientResolveRevisionFunc) bppendCbll(r0 ClientResolveRevisionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientResolveRevisionFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientResolveRevisionFunc) History() []ClientResolveRevisionFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientResolveRevisionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientResolveRevisionFuncCbll is bn object thbt describes bn invocbtion
// of method ResolveRevision on bn instbnce of MockClient.
type ClientResolveRevisionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 ResolveRevisionOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bpi.CommitID
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientResolveRevisionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientResolveRevisionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientResolveRevisionsFunc describes the behbvior when the
// ResolveRevisions method of the pbrent MockClient instbnce is invoked.
type ClientResolveRevisionsFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)
	hooks       []func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)
	history     []ClientResolveRevisionsFuncCbll
	mutex       sync.Mutex
}

// ResolveRevisions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) ResolveRevisions(v0 context.Context, v1 bpi.RepoNbme, v2 []protocol.RevisionSpecifier) ([]string, error) {
	r0, r1 := m.ResolveRevisionsFunc.nextHook()(v0, v1, v2)
	m.ResolveRevisionsFunc.bppendCbll(ClientResolveRevisionsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveRevisions
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientResolveRevisionsFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveRevisions method of the pbrent MockClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *ClientResolveRevisionsFunc) PushHook(hook func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientResolveRevisionsFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientResolveRevisionsFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
		return r0, r1
	})
}

func (f *ClientResolveRevisionsFunc) nextHook() func(context.Context, bpi.RepoNbme, []protocol.RevisionSpecifier) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientResolveRevisionsFunc) bppendCbll(r0 ClientResolveRevisionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientResolveRevisionsFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientResolveRevisionsFunc) History() []ClientResolveRevisionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientResolveRevisionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientResolveRevisionsFuncCbll is bn object thbt describes bn invocbtion
// of method ResolveRevisions on bn instbnce of MockClient.
type ClientResolveRevisionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []protocol.RevisionSpecifier
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientResolveRevisionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientResolveRevisionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientRevListFunc describes the behbvior when the RevList method of the
// pbrent MockClient instbnce is invoked.
type ClientRevListFunc struct {
	defbultHook func(context.Context, string, string, func(commit string) (bool, error)) error
	hooks       []func(context.Context, string, string, func(commit string) (bool, error)) error
	history     []ClientRevListFuncCbll
	mutex       sync.Mutex
}

// RevList delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) RevList(v0 context.Context, v1 string, v2 string, v3 func(commit string) (bool, error)) error {
	r0 := m.RevListFunc.nextHook()(v0, v1, v2, v3)
	m.RevListFunc.bppendCbll(ClientRevListFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the RevList method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientRevListFunc) SetDefbultHook(hook func(context.Context, string, string, func(commit string) (bool, error)) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RevList method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientRevListFunc) PushHook(hook func(context.Context, string, string, func(commit string) (bool, error)) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientRevListFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, string, func(commit string) (bool, error)) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientRevListFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, string, func(commit string) (bool, error)) error {
		return r0
	})
}

func (f *ClientRevListFunc) nextHook() func(context.Context, string, string, func(commit string) (bool, error)) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientRevListFunc) bppendCbll(r0 ClientRevListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientRevListFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientRevListFunc) History() []ClientRevListFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientRevListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientRevListFuncCbll is bn object thbt describes bn invocbtion of method
// RevList on bn instbnce of MockClient.
type ClientRevListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 func(commit string) (bool, error)
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientRevListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientRevListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// ClientSebrchFunc describes the behbvior when the Sebrch method of the
// pbrent MockClient instbnce is invoked.
type ClientSebrchFunc struct {
	defbultHook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)
	hooks       []func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)
	history     []ClientSebrchFuncCbll
	mutex       sync.Mutex
}

// Sebrch delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Sebrch(v0 context.Context, v1 *protocol.SebrchRequest, v2 func([]protocol.CommitMbtch)) (bool, error) {
	r0, r1 := m.SebrchFunc.nextHook()(v0, v1, v2)
	m.SebrchFunc.bppendCbll(ClientSebrchFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Sebrch method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientSebrchFunc) SetDefbultHook(hook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Sebrch method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientSebrchFunc) PushHook(hook func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientSebrchFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientSebrchFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
		return r0, r1
	})
}

func (f *ClientSebrchFunc) nextHook() func(context.Context, *protocol.SebrchRequest, func([]protocol.CommitMbtch)) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientSebrchFunc) bppendCbll(r0 ClientSebrchFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientSebrchFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientSebrchFunc) History() []ClientSebrchFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientSebrchFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientSebrchFuncCbll is bn object thbt describes bn invocbtion of method
// Sebrch on bn instbnce of MockClient.
type ClientSebrchFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *protocol.SebrchRequest
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 func([]protocol.CommitMbtch)
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientSebrchFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientSebrchFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientStbtFunc describes the behbvior when the Stbt method of the pbrent
// MockClient instbnce is invoked.
type ClientStbtFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)
	history     []ClientStbtFuncCbll
	mutex       sync.Mutex
}

// Stbt delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) Stbt(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 bpi.CommitID, v4 string) (fs.FileInfo, error) {
	r0, r1 := m.StbtFunc.nextHook()(v0, v1, v2, v3, v4)
	m.StbtFunc.bppendCbll(ClientStbtFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Stbt method of the
// pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientStbtFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Stbt method of the pbrent MockClient instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *ClientStbtFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientStbtFunc) SetDefbultReturn(r0 fs.FileInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientStbtFunc) PushReturn(r0 fs.FileInfo, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
		return r0, r1
	})
}

func (f *ClientStbtFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (fs.FileInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientStbtFunc) bppendCbll(r0 ClientStbtFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientStbtFuncCbll objects describing the
// invocbtions of this function.
func (f *ClientStbtFunc) History() []ClientStbtFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientStbtFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientStbtFuncCbll is bn object thbt describes bn invocbtion of method
// Stbt on bn instbnce of MockClient.
type ClientStbtFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bpi.CommitID
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 fs.FileInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientStbtFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientStbtFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientStrebmBlbmeFileFunc describes the behbvior when the StrebmBlbmeFile
// method of the pbrent MockClient instbnce is invoked.
type ClientStrebmBlbmeFileFunc struct {
	defbultHook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error)
	hooks       []func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error)
	history     []ClientStrebmBlbmeFileFuncCbll
	mutex       sync.Mutex
}

// StrebmBlbmeFile delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) StrebmBlbmeFile(v0 context.Context, v1 buthz.SubRepoPermissionChecker, v2 bpi.RepoNbme, v3 string, v4 *BlbmeOptions) (HunkRebder, error) {
	r0, r1 := m.StrebmBlbmeFileFunc.nextHook()(v0, v1, v2, v3, v4)
	m.StrebmBlbmeFileFunc.bppendCbll(ClientStrebmBlbmeFileFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the StrebmBlbmeFile
// method of the pbrent MockClient instbnce is invoked bnd the hook queue is
// empty.
func (f *ClientStrebmBlbmeFileFunc) SetDefbultHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StrebmBlbmeFile method of the pbrent MockClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientStrebmBlbmeFileFunc) PushHook(hook func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientStrebmBlbmeFileFunc) SetDefbultReturn(r0 HunkRebder, r1 error) {
	f.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientStrebmBlbmeFileFunc) PushReturn(r0 HunkRebder, r1 error) {
	f.PushHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error) {
		return r0, r1
	})
}

func (f *ClientStrebmBlbmeFileFunc) nextHook() func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string, *BlbmeOptions) (HunkRebder, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientStrebmBlbmeFileFunc) bppendCbll(r0 ClientStrebmBlbmeFileFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientStrebmBlbmeFileFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientStrebmBlbmeFileFunc) History() []ClientStrebmBlbmeFileFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientStrebmBlbmeFileFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientStrebmBlbmeFileFuncCbll is bn object thbt describes bn invocbtion
// of method StrebmBlbmeFile on bn instbnce of MockClient.
type ClientStrebmBlbmeFileFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 buthz.SubRepoPermissionChecker
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.RepoNbme
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 *BlbmeOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 HunkRebder
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientStrebmBlbmeFileFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientStrebmBlbmeFileFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientSystemInfoFunc describes the behbvior when the SystemInfo method of
// the pbrent MockClient instbnce is invoked.
type ClientSystemInfoFunc struct {
	defbultHook func(context.Context, string) (SystemInfo, error)
	hooks       []func(context.Context, string) (SystemInfo, error)
	history     []ClientSystemInfoFuncCbll
	mutex       sync.Mutex
}

// SystemInfo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) SystemInfo(v0 context.Context, v1 string) (SystemInfo, error) {
	r0, r1 := m.SystemInfoFunc.nextHook()(v0, v1)
	m.SystemInfoFunc.bppendCbll(ClientSystemInfoFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SystemInfo method of
// the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientSystemInfoFunc) SetDefbultHook(hook func(context.Context, string) (SystemInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SystemInfo method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientSystemInfoFunc) PushHook(hook func(context.Context, string) (SystemInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientSystemInfoFunc) SetDefbultReturn(r0 SystemInfo, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (SystemInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientSystemInfoFunc) PushReturn(r0 SystemInfo, r1 error) {
	f.PushHook(func(context.Context, string) (SystemInfo, error) {
		return r0, r1
	})
}

func (f *ClientSystemInfoFunc) nextHook() func(context.Context, string) (SystemInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientSystemInfoFunc) bppendCbll(r0 ClientSystemInfoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientSystemInfoFuncCbll objects describing
// the invocbtions of this function.
func (f *ClientSystemInfoFunc) History() []ClientSystemInfoFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientSystemInfoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientSystemInfoFuncCbll is bn object thbt describes bn invocbtion of
// method SystemInfo on bn instbnce of MockClient.
type ClientSystemInfoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 SystemInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientSystemInfoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientSystemInfoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// ClientSystemsInfoFunc describes the behbvior when the SystemsInfo method
// of the pbrent MockClient instbnce is invoked.
type ClientSystemsInfoFunc struct {
	defbultHook func(context.Context) ([]SystemInfo, error)
	hooks       []func(context.Context) ([]SystemInfo, error)
	history     []ClientSystemsInfoFuncCbll
	mutex       sync.Mutex
}

// SystemsInfo delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) SystemsInfo(v0 context.Context) ([]SystemInfo, error) {
	r0, r1 := m.SystemsInfoFunc.nextHook()(v0)
	m.SystemsInfoFunc.bppendCbll(ClientSystemsInfoFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SystemsInfo method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientSystemsInfoFunc) SetDefbultHook(hook func(context.Context) ([]SystemInfo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SystemsInfo method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientSystemsInfoFunc) PushHook(hook func(context.Context) ([]SystemInfo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientSystemsInfoFunc) SetDefbultReturn(r0 []SystemInfo, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]SystemInfo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientSystemsInfoFunc) PushReturn(r0 []SystemInfo, r1 error) {
	f.PushHook(func(context.Context) ([]SystemInfo, error) {
		return r0, r1
	})
}

func (f *ClientSystemsInfoFunc) nextHook() func(context.Context) ([]SystemInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientSystemsInfoFunc) bppendCbll(r0 ClientSystemsInfoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientSystemsInfoFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientSystemsInfoFunc) History() []ClientSystemsInfoFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientSystemsInfoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientSystemsInfoFuncCbll is bn object thbt describes bn invocbtion of
// method SystemsInfo on bn instbnce of MockClient.
type ClientSystemsInfoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []SystemInfo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientSystemsInfoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientSystemsInfoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
