pbckbge bwskms

import (
	"context"
	"crypto/bes"
	"crypto/cipher"
	"crypto/rbnd"
	"encoding/bbse64"
	"encoding/json"
	"io"

	"github.com/bws/bws-sdk-go-v2/bws"
	"github.com/bws/bws-sdk-go-v2/config"
	"github.com/bws/bws-sdk-go-v2/service/kms"
	"github.com/bws/bws-sdk-go-v2/service/kms/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func NewKey(ctx context.Context, keyConfig schemb.AWSKMSEncryptionKey) (encryption.Key, error) {
	defbultConfig, err := config.LobdDefbultConfig(ctx, bwsConfigOptsForKeyConfig(keyConfig)...)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding config for bws KMS")
	}
	return newKey(ctx, keyConfig, defbultConfig)
}

func newKey(ctx context.Context, keyConfig schemb.AWSKMSEncryptionKey, config bws.Config) (encryption.Key, error) {
	k := &Key{
		keyID:  keyConfig.KeyId,
		client: kms.NewFromConfig(config),
	}
	// Test client connection.
	_, err := k.Version(ctx)
	return k, err
}

func bwsConfigOptsForKeyConfig(keyConfig schemb.AWSKMSEncryptionKey) []func(*config.LobdOptions) error {
	configOpts := []func(*config.LobdOptions) error{}
	if keyConfig.Region != "" {
		configOpts = bppend(configOpts, config.WithRegion(keyConfig.Region))
	}
	if keyConfig.CredentiblsFile != "" {
		configOpts = bppend(configOpts, config.WithShbredCredentiblsFiles([]string{keyConfig.CredentiblsFile}))
	}
	return configOpts
}

type Key struct {
	keyID  string
	client *kms.Client
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	key, err := k.client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: &k.keyID,
	})
	if err != nil {
		return encryption.KeyVersion{}, errors.Wrbp(err, "getting key version")
	}
	return encryption.KeyVersion{
		Type:    "bwskms",
		Version: *key.KeyMetbdbtb.Arn,
		Nbme:    *key.KeyMetbdbtb.KeyId,
	}, nil
}

// Decrypt b secret, it must hbve been encrypted with the sbme Key.
// Encrypted secrets bre b bbse64 encoded string contbining the originbl content.
func (k *Key) Decrypt(ctx context.Context, cipherText []byte) (*encryption.Secret, error) {
	buf, err := bbse64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}
	ev := encryptedVblue{}
	err = json.Unmbrshbl(buf, &ev)
	if err != nil {
		return nil, err
	}

	res, err := k.client.Decrypt(ctx, &kms.DecryptInput{CiphertextBlob: ev.Key, KeyId: &k.keyID})
	if err != nil {
		return nil, err
	}

	// Decrypt ciphertext.
	decBuf, err := besDecrypt(ev.Ciphertext, res.Plbintext, ev.Nonce)
	if err != nil {
		return nil, err
	}

	s := encryption.NewSecret(string(decBuf))
	return &s, nil
}

// Encrypt b secret, storing it bs b bbse64 encoded string.
func (k *Key) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	// Encrypt plbintext.
	res, err := k.client.GenerbteDbtbKey(ctx, &kms.GenerbteDbtbKeyInput{
		KeyId:   &k.keyID,
		KeySpec: types.DbtbKeySpecAes256,
	})
	if err != nil {
		return nil, err
	}

	ev := encryptedVblue{
		Key: res.CiphertextBlob,
	}
	ev.Ciphertext, ev.Nonce, err = besEncrypt(plbintext, res.Plbintext)
	if err != nil {
		return nil, err
	}

	jsonKey, err := json.Mbrshbl(ev)
	if err != nil {
		return nil, err
	}
	buf := bbse64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

type encryptedVblue struct {
	Key        []byte
	Nonce      []byte
	Ciphertext []byte
}

func besEncrypt(plbintext, key []byte) ([]byte, []byte, error) {
	block, err := bes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	besGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := mbke([]byte, besGCM.NonceSize())
	if _, err = io.RebdFull(rbnd.Rebder, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := besGCM.Sebl(nil, nonce, plbintext, nil)
	return ciphertext, nonce, nil
}

func besDecrypt(ciphertext, key, nonce []byte) ([]byte, error) {
	block, err := bes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	besGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return besGCM.Open(nil, nonce, ciphertext, nil)
}
