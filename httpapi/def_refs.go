package httpapi

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

type file struct {
	Name     string
	RefCount int
}

type fileList []file

func (f fileList) Len() int           { return len(f) }
func (f fileList) Less(i, j int) bool { return f[i].RefCount < f[j].RefCount }
func (f fileList) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func sortByRefCount(refsPerFile map[string]int) fileList {
	fl := make(fileList, len(refsPerFile))
	i := 0
	for k, v := range refsPerFile {
		fl[i] = file{k, v}
		i++
	}
	sort.Sort(sort.Reverse(fl))
	return fl
}

func serveDefRefs(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, _, _, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     def.Repo,
		CommitID: def.CommitID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 10000
	}
	opt.Repo = defSpec.Repo

	// TODO We need a more reliable method to fetch an index of all refs by file.
	refs, err := cl.Defs.ListRefs(ctx, &sourcegraph.DefsListRefsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	refsPerFile := make(map[string]int)
	for _, ref := range refs.Refs {
		refsPerFile[ref.File]++
	}
	fl := sortByRefCount(refsPerFile)

	return writeJSON(w, &struct {
		Total int
		Files fileList
		Refs  []*graph.Ref
	}{
		Total: len(refs.Refs),
		Files: fl,
		Refs:  refs.Refs,
	})
}
