package golang_def

import (
	"encoding/json"
	"path"

	"strings"

	"sourcegraph.com/sourcegraph/srclib-go/gog/definfo"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// DefData is extra Go-specific data about a def.
type DefData struct {
	definfo.DefInfo

	// PackageImportPath is the import path of the package containing this
	// def (if this def is not a package). If this def is a package,
	// PackageImportPath is its own import path.
	PackageImportPath string `json:",omitempty"`
}

func init() {
	graph.RegisterMakeDefFormatter("GoPackage", newDefFormatter)
}

func newDefFormatter(s *graph.Def) graph.DefFormatter {
	var si DefData
	if len(s.Data) > 0 {
		if err := json.Unmarshal(s.Data, &si); err != nil {
			panic("unmarshal Go def data: " + err.Error())
		}
	}
	return defFormatter{s, &si}
}

type defFormatter struct {
	def  *graph.Def
	info *DefData
}

func (f defFormatter) Language() string { return "Go" }

func (f defFormatter) DefKeyword() string {
	switch f.info.Kind {
	case definfo.Func:
		return "func"
	case definfo.Var:
		if f.info.FieldOfStruct == "" && f.info.PkgScope {
			return "var"
		}
	case definfo.Type:
		return "type"
	case definfo.Package:
		return "package"
	case definfo.Interface:
		return "interface"
	case definfo.Const:
		return "const"
	}
	return ""
}

func (f defFormatter) Kind() string { return f.info.Kind }

func (f defFormatter) pkgPath(qual graph.Qualification) string {
	switch qual {
	case graph.DepQualified:
		return f.info.PkgName
	case graph.RepositoryWideQualified:
		// keep the last path component from the repo
		relPkg := strings.TrimPrefix(strings.TrimPrefix(f.info.PackageImportPath, f.def.Repo), "/")
		if relPkg == "" {
			relPkg = f.info.PkgName
		}
		return relPkg
	case graph.LanguageWideQualified:
		return f.info.PackageImportPath
	}
	return ""
}

func (f defFormatter) Name(qual graph.Qualification) string {
	if qual == graph.Unqualified {
		return f.def.Name
	}

	var recvlike string
	if f.info.Kind == definfo.Field {
		recvlike = f.info.FieldOfStruct
	} else if f.info.Kind == definfo.Method {
		recvlike = f.info.Receiver
	}

	pkg := f.pkgPath(qual)

	if f.info.Kind == definfo.Package {
		if qual == graph.ScopeQualified {
			pkg = f.def.Name // otherwise it'd be empty
		}
		return pkg
	}

	var prefix string
	if recvlike != "" {
		prefix = fmtReceiver(recvlike, pkg)
	} else if pkg != "" {
		prefix = pkg + "."
	}

	return prefix + f.def.Name
}

// fmtReceiver formats strings like `(*a/b.T).`.
func fmtReceiver(recv string, pkg string) string {
	// deref recv
	var recvName, ptrs string
	if i := strings.LastIndex(recv, "*"); i > -1 {
		ptrs = recv[:i+1]
		recvName = recv[i+1:]
	} else {
		recvName = recv
	}

	if pkg != "" {
		pkg += "."
	}

	return "(" + ptrs + pkg + recvName + ")."
}

func (f defFormatter) NameAndTypeSeparator() string {
	if f.info.Kind == definfo.Func || f.info.Kind == definfo.Method {
		return ""
	}
	return " "
}

func (f defFormatter) Type(qual graph.Qualification) string {
	var ts string
	switch f.def.Kind {
	case "func":
		ts = f.info.TypeString
		ts = strings.TrimPrefix(ts, "func")
	case "type":
		ts = f.info.UnderlyingTypeString
		if i := strings.Index(ts, "{"); i != -1 {
			ts = ts[:i]
		}
		ts = " " + ts
	default:
		ts = " " + f.info.TypeString
	}

	// qualify the package path based on qual
	oldPkgPath := f.info.PackageImportPath + "."
	newPkgPath := f.pkgPath(qual)
	if newPkgPath != "" {
		newPkgPath += "."
	}
	ts = strings.Replace(ts, oldPkgPath, newPkgPath, -1)

	ts = strings.Replace(ts, f.def.Repo+"/", "", -1)
	ts = strings.Replace(ts, f.def.Repo+".", path.Base(f.def.Repo), -1)

	return ts
}
