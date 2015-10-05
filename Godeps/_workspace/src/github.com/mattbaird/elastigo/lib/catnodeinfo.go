package elastigo

import (
	"fmt"
	"strconv"
	"strings"
)



// newCatNodeInfo returns an instance of CatNodeInfo populated with the
// the information in the cat output indexLine which contains the
// specified fields. An err is returned if a field is not known.
func newCatNodeInfo(fields []string, indexLine string) (catNode *CatNodeInfo, err error) {

	split := strings.Fields(indexLine)
	catNode = &CatNodeInfo{}

	// Check the fields length compared to the number of stats
	lf, ls := len(fields), len(split)
	if lf > ls {
		return nil, fmt.Errorf("Number of fields (%d) greater than number of stats (%d)", lf, ls)
	}

	// Populate the apropriate field in CatNodeInfo
	for i, field := range fields {

		switch field {
		case "id", "nodeId":
			catNode.Id = split[i]
		case "pid", "p":
			catNode.PID = split[i]
		case "host", "h":
			catNode.Host = split[i]
		case "ip", "i":
			catNode.IP = split[i]
		case "port", "po":
			catNode.Port = split[i]
		case "version", "v":
			catNode.Version = split[i]
		case "build", "b":
			catNode.Build = split[i]
		case "jdk", "j":
			catNode.JDK = split[i]
		case "disk.avail", "d", "disk", "diskAvail":
			catNode.DiskAvail = split[i]
		case "heap.current", "hc", "heapCurrent":
			catNode.HeapCur = split[i]
		case "heap.percent", "hp", "heapPercent":
			catNode.HeapPerc = split[i]
		case "heap.max", "hm", "heapMax":
			catNode.HeapMax = split[i]
		case "ram.current", "rc", "ramCurrent":
			catNode.RamCur = split[i]
		case "ram.percent", "rp", "ramPercent":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.RamPerc = int16(val)
		case "ram.max", "rm", "ramMax":
			catNode.RamMax = split[i]
		case "file_desc.current", "fdc", "fileDescriptorCurrent":
			catNode.FileDescCur = split[i]
		case "file_desc.percent", "fdp", "fileDescriptorPercent":
			catNode.FileDescPerc = split[i]
		case "file_desc.max", "fdm", "fileDescriptorMax":
			catNode.FileDescMax = split[i]
		case "load", "l":
			catNode.Load = split[i]
		case "uptime", "u":
			catNode.UpTime = split[i]
		case "node.role", "r", "role", "dc", "nodeRole":
			catNode.NodeRole = split[i]
		case "master", "m":
			catNode.Master = split[i]
		case "name", "n":
			catNode.Name = strings.Join(split[i:], " ")
		case "completion.size", "cs", "completionSize":
			catNode.CmpltSize = split[i]
		case "fielddata.memory_size", "fm", "fielddataMemory":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.FieldMem = val
		case "fielddata.evictions", "fe", "fieldataEvictions":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.FieldEvict = val
		case "filter_cache.memory_size", "fcm", "filterCacheMemory":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.FiltMem = val
		case "filter_cache.evictions", "fce", "filterCacheEvictions":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.FiltEvict = val
		case "flush.total", "ft", "flushTotal":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.FlushTotal = val
		case "flush.total_time", "ftt", "flushTotalTime":
			catNode.FlushTotalTime = split[i]
		case "get.current", "gc", "getCurrent":
			catNode.GetCur = split[i]
		case "get.time", "gti", "getTime":
			catNode.GetTime = split[i]
		case "get.total", "gto", "getTotal":
			catNode.GetTotal = split[i]
		case "get.exists_time", "geti", "getExistsTime":
			catNode.GetExistsTime = split[i]
		case "get.exists_total", "geto", "getExistsTotal":
			catNode.GetExistsTotal = split[i]
		case "get.missing_time", "gmti", "getMissingTime":
			catNode.GetMissingTime = split[i]
		case "get.missing_total", "gmto", "getMissingTotal":
			catNode.GetMissingTotal = split[i]
		case "id_cache.memory_size", "im", "idCacheMemory":
			val, err := strconv.Atoi(split[i])
			if err != nil {
				return nil, err
			}
			catNode.IDCacheMemory = val
		case "indexing.delete_current", "idc", "indexingDeleteCurrent":
			catNode.IdxDelCur = split[i]
		case "indexing.delete_time", "idti", "indexingDeleteime":
			catNode.IdxDelTime = split[i]
		case "indexing.delete_total", "idto", "indexingDeleteTotal":
			catNode.IdxDelTotal = split[i]
		case "indexing.index_current", "iic", "indexingIndexCurrent":
			catNode.IdxIdxCur = split[i]
		case "indexing.index_time", "iiti", "indexingIndexTime":
			catNode.IdxIdxTime = split[i]
		case "indexing.index_total", "iito", "indexingIndexTotal":
			catNode.IdxIdxTotal = split[i]
		case "merges.current", "mc", "mergesCurrent":
			catNode.MergCur = split[i]
		case "merges.current_docs", "mcd", "mergesCurrentDocs":
			catNode.MergCurDocs = split[i]
		case "merges.current_size", "mcs", "mergesCurrentSize":
			catNode.MergCurSize = split[i]
		case "merges.total", "mt", "mergesTotal":
			catNode.MergTotal = split[i]
		case "merges.total_docs", "mtd", "mergesTotalDocs":
			catNode.MergTotalDocs = split[i]
		case "merges.total_size", "mts", "mergesTotalSize":
			catNode.MergTotalSize = split[i]
		case "merges.total_time", "mtt", "mergesTotalTime":
			catNode.MergTotalTime = split[i]
		case "percolate.current", "pc", "percolateCurrent":
			catNode.PercCur = split[i]
		case "percolate.memory_size", "pm", "percolateMemory":
			catNode.PercMem = split[i]
		case "percolate.queries", "pq", "percolateQueries":
			catNode.PercQueries = split[i]
		case "percolate.time", "pti", "percolateTime":
			catNode.PercTime = split[i]
		case "percolate.total", "pto", "percolateTotal":
			catNode.PercTotal = split[i]
		case "refesh.total", "rto", "refreshTotal":
			catNode.RefreshTotal = split[i]
		case "refresh.time", "rti", "refreshTime":
			catNode.RefreshTime = split[i]
		case "search.fetch_current", "sfc", "searchFetchCurrent":
			catNode.SearchFetchCur = split[i]
		case "search.fetch_time", "sfti", "searchFetchTime":
			catNode.SearchFetchTime = split[i]
		case "search.fetch_total", "sfto", "searchFetchTotal":
			catNode.SearchFetchTotal = split[i]
		case "search.open_contexts", "so", "searchOpenContexts":
			catNode.SearchOpenContexts = split[i]
		case "search.query_current", "sqc", "searchQueryCurrent":
			catNode.SearchQueryCur = split[i]
		case "search.query_time", "sqti", "searchQueryTime":
			catNode.SearchQueryTime = split[i]
		case "search.query_total", "sqto", "searchQueryTotal":
			catNode.SearchQueryTotal = split[i]
		case "segments.count", "sc", "segmentsCount":
			catNode.SegCount = split[i]
		case "segments.memory", "sm", "segmentsMemory":
			catNode.SegMem = split[i]
		case "segments.index_writer_memory", "siwm", "segmentsIndexWriterMemory":
			catNode.SegIdxWriterMem = split[i]
		case "segments.index_writer_max_memory", "siwmx", "segmentsIndexWriterMaxMemory":
			catNode.SegIdxWriterMax = split[i]
		case "segments.version_map_memory", "svmm", "segmentsVersionMapMemory":
			catNode.SegVerMapMem = split[i]
		default:
			return nil, fmt.Errorf("Invalid cat nodes field: %s", field)
		}
	}

	return catNode, nil
}

// GetCatNodeInfo issues an elasticsearch cat nodes request with the specified
// fields and returns a list of CatNodeInfos, one for each node, whose requested
// members are populated with statistics. If fields is nil or empty, the default
// cat output is used.
// NOTE: if you include the name field, make sure it is the last field in the
// list, because name values can contain spaces which screw up the parsing
func (c *Conn) GetCatNodeInfo(fields []string) (catNodes []CatNodeInfo, err error) {

	catNodes = make([]CatNodeInfo, 0)

	// If no fields have been specified, use the "default" arrangement
	if len(fields) < 1 {
		fields = []string{"host", "ip", "heap.percent", "ram.percent", "load",
			"node.role", "master", "name"}
	}

	// Issue a request for stats on the requested fields
	args := map[string]interface{}{
		"bytes": "b",
		"h":     strings.Join(fields, ","),
	}
	indices, err := c.DoCommand("GET", "/_cat/nodes/", args, nil)
	if err != nil {
		return catNodes, err
	}

	// Create a CatIndexInfo for each line in the response
	indexLines := strings.Split(string(indices[:]), "\n")
	for _, index := range indexLines {

		// Ignore empty output lines
		if len(index) < 1 {
			continue
		}

		// Create a CatNodeInfo and append it to the result
		info, err := newCatNodeInfo(fields, index)
		if info != nil {
			catNodes = append(catNodes, *info)
		} else if err != nil {
			return catNodes, err
		}
	}
	return catNodes, nil
}
