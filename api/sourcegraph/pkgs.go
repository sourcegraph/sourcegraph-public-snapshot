package sourcegraph

type PackageInfo struct {
	RepoID int32
	Lang   string
	Pkg    map[string]interface{}
}

type ListPackagesOp struct {
	Lang     string
	PkgQuery map[string]interface{}
	Limit    int
}
