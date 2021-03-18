package conversion

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

func newID() (semantic.ID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return semantic.ID(uuid.String()), nil
}

func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func toID(id int) semantic.ID {
	if id == 0 {
		return semantic.ID("")
	}

	return semantic.ID(strconv.FormatInt(int64(id), 10))
}
