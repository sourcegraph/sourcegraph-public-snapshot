package conversion

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func newID() (precise.ID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return precise.ID(uuid.String()), nil
}

func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}

func toID(id int) precise.ID {
	if id == 0 {
		return precise.ID("")
	}

	return precise.ID(strconv.FormatInt(int64(id), 10))
}
