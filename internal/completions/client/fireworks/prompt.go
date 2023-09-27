pbckbge fireworks

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func getPrompt(messbges []types.Messbge) (string, error) {
	if len(messbges) != 1 {
		return "", errors.New("Expected to receive exbctly one messbge with the prompt")
	}

	return messbges[0].Text, nil
}
