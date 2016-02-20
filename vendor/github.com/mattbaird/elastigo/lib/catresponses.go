package elastigo

type CatIndexInfo struct {
	Health   string
	Status   string
	Name     string
	Shards   int
	Replicas int
	Docs     CatIndexDocs
	Store    CatIndexStore
}

type CatIndexDocs struct {
	Count   int64
	Deleted int64
}

type CatIndexStore struct {
	Size    int64
	PriSize int64
}

type CatShardInfo struct {
	IndexName string
	Shard     int
	Primary   string
	State     string
	Docs      int64
	Store     int64
	NodeIP    string
	NodeName  string
}

type CatNodeInfo struct {
	Id                 string
	PID                string
	Host               string
	IP                 string
	Port               string
	Version            string
	Build              string
	JDK                string
	DiskAvail          string
	HeapCur            string
	HeapPerc           string
	HeapMax            string
	RamCur             string
	RamPerc            int16
	RamMax             string
	FileDescCur        string
	FileDescPerc       string
	FileDescMax        string
	Load               string
	UpTime             string
	NodeRole           string
	Master             string
	Name               string
	CmpltSize          string
	FieldMem           int
	FieldEvict         int
	FiltMem            int
	FiltEvict          int
	FlushTotal         int
	FlushTotalTime     string
	GetCur             string
	GetTime            string
	GetTotal           string
	GetExistsTime      string
	GetExistsTotal     string
	GetMissingTime     string
	GetMissingTotal    string
	IDCacheMemory      int
	IdxDelCur          string
	IdxDelTime         string
	IdxDelTotal        string
	IdxIdxCur          string
	IdxIdxTime         string
	IdxIdxTotal        string
	MergCur            string
	MergCurDocs        string
	MergCurSize        string
	MergTotal          string
	MergTotalDocs      string
	MergTotalSize      string
	MergTotalTime      string
	PercCur            string
	PercMem            string
	PercQueries        string
	PercTime           string
	PercTotal          string
	RefreshTotal       string
	RefreshTime        string
	SearchFetchCur     string
	SearchFetchTime    string
	SearchFetchTotal   string
	SearchOpenContexts string
	SearchQueryCur     string
	SearchQueryTime    string
	SearchQueryTotal   string
	SegCount           string
	SegMem             string
	SegIdxWriterMem    string
	SegIdxWriterMax    string
	SegVerMapMem       string
}
