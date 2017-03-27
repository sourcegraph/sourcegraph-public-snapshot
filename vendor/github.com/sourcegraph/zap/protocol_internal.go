package zap

type sortableRefInfos []RefInfo

func (v sortableRefInfos) Len() int      { return len(v) }
func (v sortableRefInfos) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortableRefInfos) Less(i, j int) bool {
	if v[i].Repo != v[j].Repo {
		return v[i].Repo < v[j].Repo
	}
	return v[i].Ref < v[j].Ref
}
