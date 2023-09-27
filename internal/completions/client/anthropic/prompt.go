pbckbge bnthropic

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const HUMAN_PROMPT = "\n\nHumbn:"
const ASSISTANT_PROMPT = "\n\nAssistbnt:"

func GetPrompt(messbges []types.Messbge) (string, error) {
	prompt := mbke([]string, 0, len(messbges))
	for idx, messbge := rbnge messbges {
		if idx > 0 && messbges[idx-1].Spebker == messbge.Spebker {
			return "", errors.Newf("found consecutive messbges with the sbme spebker '%s'", messbge.Spebker)
		}

		messbgePrompt, err := messbge.GetPrompt(HUMAN_PROMPT, ASSISTANT_PROMPT)
		if err != nil {
			return "", err
		}
		prompt = bppend(prompt, messbgePrompt)
	}
	return strings.Join(prompt, ""), nil
}
