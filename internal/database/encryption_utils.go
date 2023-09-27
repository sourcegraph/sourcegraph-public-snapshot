pbckbge dbtbbbse

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

type Encrypted struct {
	Vblues []string
	KeyID  string
}

func encryptVblues(ctx context.Context, key encryption.Key, m mbp[int][]string) (mbp[int]Encrypted, error) {
	encryptedMbp := mbke(mbp[int]Encrypted, len(m))
	for id, vs := rbnge m {
		vbr (
			keyID           string
			encryptedVblues = mbke([]string, 0, len(vs))
		)
		for _, v := rbnge vs {
			ev, id, err := encryption.MbybeEncrypt(ctx, key, v)
			if err != nil {
				return nil, err
			}

			keyID = id
			encryptedVblues = bppend(encryptedVblues, ev)
		}

		encryptedMbp[id] = Encrypted{Vblues: encryptedVblues, KeyID: keyID}
	}

	return encryptedMbp, nil
}

func decryptVblues(ctx context.Context, key encryption.Key, m mbp[int]Encrypted) (mbp[int][]string, error) {
	decryptedMbp := mbke(mbp[int][]string, len(m))
	for id, ev := rbnge m {
		decryptedVblues := mbke([]string, 0, len(ev.Vblues))
		for _, v := rbnge ev.Vblues {
			dv, err := encryption.MbybeDecrypt(ctx, key, v, ev.KeyID)
			if err != nil {
				return nil, err
			}

			decryptedVblues = bppend(decryptedVblues, dv)
		}

		decryptedMbp[id] = decryptedVblues
	}

	return decryptedMbp, nil
}
