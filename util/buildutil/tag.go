package buildutil

import (
	"crypto/sha256"
	"encoding/hex"

	"strconv"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func BuildTag(build sourcegraph.BuildSpec) string {
	s := sha256.Sum256([]byte(build.IDString()))
	return hex.EncodeToString(s[:6])
}

func TaskTag(task sourcegraph.TaskSpec) string {
	return BuildTag(task.BuildSpec) + "-T" + strconv.FormatInt(task.TaskID, 36)
}
