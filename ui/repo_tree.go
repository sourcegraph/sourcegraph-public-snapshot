package ui

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/ui/payloads"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	opt := &sourcegraph.RepoTreeGetOptions{}
	if err := schemaDecoder.Decode(opt, r.URL.Query()); err != nil {
		return err
	}

	e := json.NewEncoder(w)
	tc, rc, vc, err := handlerutil.GetTreeEntryCommon(ctx, mux.Vars(r), opt)
	if err != nil {
		if urlErr, ok := err.(*handlerutil.URLMovedError); ok {
			return e.Encode(urlErr)
		}
		return err
	}
	if err != nil {
		return err
	}

	treeSearchOpt := struct{ Recursive bool }{}
	schemaDecoder.Decode(&treeSearchOpt, r.URL.Query())
	if treeSearchOpt.Recursive {
		return e.Encode(makeFileList(tc.Entry))
	}

	tc.Entry.ContentsString = string(tc.Entry.Contents)
	tc.Entry.Contents = nil

	return e.Encode(payloads.CodeFile{
		Repo:              rc.Repo,
		RepoCommit:        vc.RepoCommit,
		EntrySpec:         tc.EntrySpec,
		SrclibDataVersion: tc.SrclibDataVersion,
		Entry:             tc.Entry,
	})
}

// makeFileList simplifies a TreeEntry to a slice of files.
func makeFileList(entry *sourcegraph.TreeEntry) []string {
	if entry == nil || entry.BasicTreeEntry == nil || entry.BasicTreeEntry.Entries == nil {
		return nil
	}
	entries := entry.BasicTreeEntry.Entries
	list := make([]string, 0, len(entries))
	for _, e := range entries {
		list = append(list, getEntries("", e)...)
	}
	return list
}

// getEntries recursively returns all files in an entry
func getEntries(prefix string, e *sourcegraph.BasicTreeEntry) []string {
	if len(e.Entries) > 0 {
		ee := make([]string, 0, len(e.Entries))
		for _, entry := range e.Entries {
			ee = append(ee, getEntries(prefix+e.Name+"/", entry)...)
		}
		return ee
	}
	return []string{prefix + e.Name}
}
