package encryption

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type Encrypted struct {
	Values []string
	KeyID  string
}

func encryptValues(ctx context.Context, key encryption.Key, m map[int][]string) (map[int]Encrypted, error) {
	encryptedMap := make(map[int]Encrypted, len(m))
	for id, vs := range m {
		var (
			keyID           string
			encryptedValues = make([]string, 0, len(vs))
		)
		for _, v := range vs {
			ev, id, err := encryption.MaybeEncrypt(ctx, key, v)
			if err != nil {
				return nil, err
			}

			keyID = id
			encryptedValues = append(encryptedValues, ev)
		}

		encryptedMap[id] = Encrypted{Values: encryptedValues, KeyID: keyID}
	}

	return encryptedMap, nil
}

func decryptValues(ctx context.Context, key encryption.Key, m map[int]Encrypted) (map[int][]string, error) {
	decryptedMap := make(map[int][]string, len(m))
	for id, ev := range m {
		decryptedValues := make([]string, 0, len(ev.Values))
		for _, v := range ev.Values {
			dv, err := encryption.MaybeDecrypt(ctx, key, v, ev.KeyID)
			if err != nil {
				return nil, err
			}

			decryptedValues = append(decryptedValues, dv)
		}

		decryptedMap[id] = decryptedValues
	}

	return decryptedMap, nil
}
