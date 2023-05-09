package resolvers

import (
	"context"
)

type ContextServiceResolver interface {
	FindMostRelevantSCIPSymbols(ctx context.Context, args *FindMostRelevantSCIPSymbolsArgs) (string, error)
}

type FindMostRelevantSCIPSymbolsArgs struct {
	Args *RelevantSCIPSymbolsArgs
}

type RelevantSCIPSymbolsArgs struct {
	// The symbol names to search for
	Symbols *[]string
	// The repository the user is in
	Repository *string
	// The commit of the repository the user is in
	CommitID *string
	// The closest remote commit of the repository the user is in
	ClosestRemoteCommitID *string
	// The state of the editor for the user
	EditorState *EditorState
}

type EditorState struct {
	// The file that is currently open in the editor
	ActiveFile *string
	// The contents of the file that is currently open in the editor
	ActiveFileContent *string
	// Whether the file that is currently open in the editor has unsaved changes
	IsActiveFileDirty *bool
	// The files that are currently open in the editor
	OpenFiles []*string
}
