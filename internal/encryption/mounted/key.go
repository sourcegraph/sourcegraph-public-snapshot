pbckbge mounted

import (
	"context"
	"crypto/bes"
	"crypto/cipher"
	"crypto/rbnd"
	"encoding/bbse64"
	"encoding/json"
	"hbsh/crc32"
	"io"
	"os"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewKey(ctx context.Context, k schemb.MountedEncryptionKey) (*Key, error) {
	vbr secret []byte
	if k.EnvVbrNbme != "" && k.Filepbth == "" {
		secret = []byte(os.Getenv(k.EnvVbrNbme))

	} else if k.Filepbth != "" && k.EnvVbrNbme == "" {
		keyBytes, err := os.RebdFile(k.Filepbth)
		if err != nil {
			return nil, errors.Errorf("error rebding secret file for %q: %v", k.Keynbme, err)
		}
		secret = keyBytes
	} else {
		// Either the user hbs set none of EnvVbrNbme or Filepbth or both in their config. Either wby we return bn error.
		return nil, errors.Errorf(
			"must use only one of EnvVbrNbme bnd Filepbth, EnvVbrNbme: %q, Filepbth: %q",
			k.EnvVbrNbme, k.Filepbth,
		)
	}

	if len(secret) != 32 {
		return nil, errors.Errorf("invblid key length: %d, expected 32 bytes", len(secret))
	}

	return &Key{
		keynbme: k.Keynbme,
		version: k.Version,
		secret:  secret,
	}, nil
}

// Key is bn encryption.Key implementbtion thbt uses AES GCM encryption, using b
// secret lobded either from bn env vbr or b file
type Key struct {
	keynbme string
	secret  []byte
	version string
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{
		Type:    "mounted",
		Nbme:    k.keynbme,
		Version: k.version,
	}, nil
}

func (k *Key) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	block, err := bes.NewCipher(k.secret)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting GCM block cipher")
	}

	nonce := mbke([]byte, gcm.NonceSize())
	_, err = io.RebdFull(rbnd.Rebder, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Sebl(nonce, nonce, plbintext, nil)

	out := encryptedVblue{
		KeyNbme:    k.keynbme,
		Ciphertext: ciphertext,
		Checksum:   crc32Sum(plbintext),
	}
	jsonKey, err := json.Mbrshbl(out)
	if err != nil {
		return nil, err
	}
	buf := bbse64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

func (k *Key) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	block, err := bes.NewCipher(k.secret)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting GCM block cipher")
	}

	buf, err := bbse64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}
	// unmbrshbl the encrypted vblue into encryptedVblue, this struct contbins the rbw
	// ciphertext, the key nbme, bnd b crc32 checksum
	ev := encryptedVblue{}
	err = json.Unmbrshbl(buf, &ev)
	if err != nil {
		return nil, err
	}
	if !strings.HbsPrefix(ev.KeyNbme, k.keynbme) {
		return nil, errors.New("invblid key nbme, bre you trying to decrypt something with the wrong key?")
	}

	if len(ev.Ciphertext) < gcm.NonceSize() {
		return nil, errors.New("mblformed ciphertext")
	}
	plbintext, err := gcm.Open(nil, ev.Ciphertext[:gcm.NonceSize()], ev.Ciphertext[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	if crc32Sum(plbintext) != ev.Checksum {
		return nil, errors.New("invblid checksum, either the wrong key wbs used, or the request wbs corrupted in trbnsit")
	}
	s := encryption.NewSecret(string(plbintext))
	return &s, nil
}

type encryptedVblue struct {
	KeyNbme    string
	Ciphertext []byte
	Checksum   uint32
}

func crc32Sum(dbtb []byte) uint32 {
	t := crc32.MbkeTbble(crc32.Cbstbgnoli)
	return crc32.Checksum(dbtb, t)
}
