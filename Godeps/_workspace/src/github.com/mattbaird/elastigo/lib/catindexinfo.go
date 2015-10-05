package elastigo

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidIndexLine = errors.New("Cannot parse indexline")

//Create an IndexInfo from the string _cat/indices would produce
//EX: health status index    pri rep docs.count docs.deleted store.size  pri.store.size
//green open logs-2015-06-19   2 0   135389346  0            53048922233 53048922233
func NewCatIndexInfo(indexLine string) (catIndex *CatIndexInfo, err error) {
	split := strings.Fields(indexLine)
	if len(split) < 5 {
		return nil, ErrInvalidIndexLine
	}
	catIndex = &CatIndexInfo{}
	catIndex.Store = CatIndexStore{}
	catIndex.Docs = CatIndexDocs{}
	catIndex.Health = split[0]
	catIndex.Status = split[1]
	catIndex.Name = split[2]
	catIndex.Shards, err = strconv.Atoi(split[3])
	if err != nil {
		catIndex.Shards = 0
	}
	catIndex.Replicas, err = strconv.Atoi(split[4])
	if err != nil {
		catIndex.Replicas = 0
	}
	if len(split) == 5 {
		return catIndex, nil
	}
	catIndex.Docs.Count, err = strconv.ParseInt(split[5], 10, 64)
	if err != nil {
		catIndex.Docs.Count = 0
	}
	if len(split) == 6 {
		return catIndex, nil
	}
	catIndex.Docs.Deleted, err = strconv.ParseInt(split[6], 10, 64)
	if err != nil {
		catIndex.Docs.Deleted = 0
	}
	if len(split) == 7 {
		return catIndex, nil
	}
	catIndex.Store.Size, err = strconv.ParseInt(split[7], 10, 64)
	if err != nil {
		catIndex.Store.Size = 0
	}
	if len(split) == 8 {
		return catIndex, nil
	}
	catIndex.Store.PriSize, err = strconv.ParseInt(split[8], 10, 64)
	if err != nil {
		catIndex.Store.PriSize = 0
	}
	return catIndex, nil
}

// Pull all the index info from the connection
func (c *Conn) GetCatIndexInfo(pattern string) (catIndices []CatIndexInfo) {
	catIndices = make([]CatIndexInfo, 0)
	//force it to only show the fileds we know about
	args := map[string]interface{}{"bytes": "b", "h": "health,status,index,pri,rep,docs.count,docs.deleted,store.size,pri.store.size"}
	indices, err := c.DoCommand("GET", "/_cat/indices/"+pattern, args, nil)
	if err == nil {
		indexLines := strings.Split(string(indices[:]), "\n")
		for _, index := range indexLines {
			ci, _ := NewCatIndexInfo(index)
			if nil != ci {
				catIndices = append(catIndices, *ci)
			}
		}
	}
	return catIndices
}
