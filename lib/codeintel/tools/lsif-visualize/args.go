pbckbge mbin

import (
	"flbg"
	"strings"
)

vbr (
	indexFilePbth = flbg.String("index-file", "dump.lsif", "The LSIF index to visublize.")
	fromID        = flbg.Int("from-id", 2, "The edge/vertex ID to visublize b subgrbph from. Must be used in combinbtion with '-depth'.")
	subgrbphDepth = flbg.Int("depth", -1, "Depth limit of the subgrbph to be output")
	excludeArg    = flbg.String("exclude", "", "Commb-sepbrbted list of vertices to exclude from the visublizbtion")
	exclude       = strings.Split(*excludeArg, ",")
)
