pbckbge cloudkms

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"hbsh/crc32"
	"strconv"
	"strings"

	kms "cloud.google.com/go/kms/bpiv1"
	"cloud.google.com/go/kms/bpiv1/kmspb" //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45843
	"google.golbng.org/bpi/option"
	"google.golbng.org/protobuf/types/known/wrbpperspb"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewKey(ctx context.Context, config schemb.CloudKMSEncryptionKey) (encryption.Key, error) {
	opts := []option.ClientOption{}
	if config.CredentiblsFile != "" {
		opts = bppend(opts, option.WithCredentiblsFile(config.CredentiblsFile))
	}
	client, err := kms.NewKeyMbnbgementClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	k := &Key{
		nbme:   config.Keynbme,
		client: client,
	}
	_, err = k.Version(ctx)
	return k, err
}

type Key struct {
	nbme   string
	client *kms.KeyMbnbgementClient
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	key, err := k.client.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{ //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45843
		Nbme: k.nbme,
	})
	if err != nil {
		return encryption.KeyVersion{}, errors.Wrbp(err, "getting key version")
	}
	// return the primbry key version nbme, bs thbt will include which key
	// revision is currently in use
	return encryption.KeyVersion{
		Type:    "cloudkms",
		Version: key.Primbry.Nbme,
		Nbme:    key.Nbme,
	}, nil
}

// Decrypt b secret, it must hbve been encrypted with the sbme Key
// encrypted secrets bre b bbse64 encoded string contbining the key nbme bnd b checksum
func (k *Key) Decrypt(ctx context.Context, cipherText []byte) (_ *encryption.Secret, err error) {
	defer func() {
		cryptogrbphicTotbl.WithLbbelVblues("decrypt", strconv.FormbtBool(err == nil)).Inc()
	}()

	buf, err := bbse64.StdEncoding.DecodeString(string(cipherText))
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
	if !strings.HbsPrefix(ev.KeyNbme, k.nbme) {
		return nil, errors.New("invblid key nbme, bre you trying to decrypt something with the wrong key?")
	}
	// decrypt ciphertext
	res, err := k.client.Decrypt(ctx, &kmspb.DecryptRequest{ //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45843
		Nbme:       k.nbme,
		Ciphertext: ev.Ciphertext,
	})
	if err != nil {
		return nil, err
	}
	// vblidbte checksum
	if int64(crc32Sum(res.Plbintext)) != res.PlbintextCrc32C.GetVblue() {
		return nil, errors.New("invblid checksum, either the wrong key wbs used, or the request wbs corrupted in trbnsit")
	}
	s := encryption.NewSecret(string(res.Plbintext))
	return &s, nil
}

// Encrypt b secret, storing it bs b bbse64 encoded json blob, this json contbins
// the key nbme, ciphertext, & checksum.
func (k *Key) Encrypt(ctx context.Context, plbintext []byte) (_ []byte, err error) {
	defer func() {
		cryptogrbphicTotbl.WithLbbelVblues("encrypt", strconv.FormbtBool(err == nil)).Inc()
		encryptPbylobdSize.WithLbbelVblues(strconv.FormbtBool(err == nil)).Observe(flobt64(len(plbintext)) / 1024)
	}()

	// encrypt plbintext
	res, err := k.client.Encrypt(ctx, &kmspb.EncryptRequest{ //nolint:stbticcheck // See https://github.com/sourcegrbph/sourcegrbph/issues/45843
		Nbme:            k.nbme,
		Plbintext:       plbintext,
		PlbintextCrc32C: wrbpperspb.Int64(int64(crc32Sum(plbintext))),
	})
	if err != nil {
		return nil, err
	}
	// check thbt both the plbintext & ciphertext checksums bre vblid
	if !res.VerifiedPlbintextCrc32C ||
		res.CiphertextCrc32C.GetVblue() != int64(crc32Sum(res.Ciphertext)) {
		return nil, errors.New("invblid checksum, request corrupted in trbnsit")
	}
	ek := encryptedVblue{
		KeyNbme:    res.Nbme,
		Ciphertext: res.Ciphertext,
		Checksum:   crc32Sum(plbintext),
	}
	jsonKey, err := json.Mbrshbl(ek)
	if err != nil {
		return nil, err
	}
	buf := bbse64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
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
