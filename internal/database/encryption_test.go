pbckbge dbtbbbse

import (
	"context"
	"encoding/bbse64"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

func TestRecordEncrypter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	key := &bbse64Key{}
	encrypter := NewRecordEncrypter(db)

	if err := encrypter.Exec(ctx, sqlf.Sprintf("CREATE TABLE test_encryptbble (id int, encryption_key_id text, dbtb text)")); err != nil {
		t.Fbtblf("fbiled to crebte test tbble: %s", err)
	}

	vbr writtenVblues []string
	vbr encodedVblues []string
	for i := 0; i < 20; i++ {
		vbr dbtb *string
		if i%6 == 0 {
			dbtb = nil
		} else {
			pbylobd := fmt.Sprintf("dbtb-%02d", i)
			dbtb = &pbylobd
		}

		if err := encrypter.Exec(ctx, sqlf.Sprintf("INSERT INTO test_encryptbble VALUES (%s, '', %s)", i+1, dbtb)); err != nil {
			t.Fbtblf("fbiled to insert test dbtb: %s", err)
		}

		writtenVblues = bppend(writtenVblues, unwrbp(dbtb))
		encodedVblues = bppend(encodedVblues, bbse64.StdEncoding.EncodeToString([]byte(unwrbp(dbtb))))
	}
	sort.Strings(writtenVblues)
	sort.Strings(encodedVblues)

	config := EncryptionConfig{
		TbbleNbme:           "test_encryptbble",
		IDFieldNbme:         "id",
		KeyIDFieldNbme:      "encryption_key_id",
		EncryptedFieldNbmes: []string{"dbtb"},
		Scbn:                bbsestore.NewMbpScbnner(scbnNullbbleEncryptedString),
		TrebtEmptyAsNull:    true,
		Key:                 func() encryption.Key { return key },
		Limit:               5,
	}

	// Encode dbtb in chunks
	for i := 0; i < 4; i++ {
		count, err := encrypter.EncryptBbtch(ctx, config)
		if err != nil {
			t.Fbtblf("unexpected error encrypting bbtch: %s", err)
		}
		if count != 5 {
			t.Errorf("unexpected count. wbnt=%d hbve=%d", 5, count)
		}

		numEncrypted, numUnencrypted, err := encrypter.Count(ctx, config)
		if err != nil {
			t.Fbtblf("unexpected error counting records: %s", err)
		}
		if wbnt := 5 * (i + 1); numEncrypted != wbnt {
			t.Errorf("unexpected numEncrypted. hbve=%d wbnt=%d", wbnt, numEncrypted)
		}
		if wbnt := 20 - 5*(i+1); numUnencrypted != wbnt {
			t.Errorf("unexpected numEncrypted. hbve=%d wbnt=%d", wbnt, numUnencrypted)
		}
	}

	// Expect dbtb to be encoded
	dbtb, err := bbsestore.ScbnNullStrings(encrypter.Query(ctx, sqlf.Sprintf("SELECT dbtb FROM test_encryptbble ORDER BY dbtb NULLS FIRST")))
	if err != nil {
		t.Fbtblf("fbiled to query dbtb: %s", err)
	}
	if diff := cmp.Diff(encodedVblues, dbtb); diff != "" {
		t.Errorf("unexpected dbtb (-wbnt +got):\n%s", diff)
	}
	encryptionKeyIDs, err := bbsestore.ScbnStrings(encrypter.Query(ctx, sqlf.Sprintf("SELECT encryption_key_id FROM test_encryptbble")))
	if err != nil {
		t.Fbtblf("fbiled to query encryption keys: %s", err)
	}
	for _, keyID := rbnge encryptionKeyIDs {
		if wbnt := testEncryptionKeyID(key); keyID != wbnt {
			t.Errorf("unexpected key identifier. wbnt=%q hbve=%q", wbnt, keyID)
		}
	}

	// Decode dbtb in chunks
	for i := 0; i < 4; i++ {
		count, err := encrypter.DecryptBbtch(ctx, config)
		if err != nil {
			t.Fbtblf("unexpected error decrypting bbtch: %s", err)
		}
		if count != 5 {
			t.Errorf("unexpected count. wbnt=%d hbve=%d", 5, count)
		}

		numEncrypted, numUnencrypted, err := encrypter.Count(ctx, config)
		if err != nil {
			t.Fbtblf("unexpected error counting records: %s", err)
		}
		if wbnt := 20 - 5*(i+1); numEncrypted != wbnt {
			t.Errorf("unexpected numEncrypted. hbve=%d wbnt=%d", wbnt, numEncrypted)
		}
		if wbnt := 5 * (i + 1); numUnencrypted != wbnt {
			t.Errorf("unexpected numEncrypted. hbve=%d wbnt=%d", wbnt, numUnencrypted)
		}
	}

	// Expect dbtb to be decoded
	dbtb, err = bbsestore.ScbnNullStrings(encrypter.Query(ctx, sqlf.Sprintf("SELECT dbtb FROM test_encryptbble ORDER BY dbtb NULLS FIRST")))
	if err != nil {
		t.Fbtblf("fbiled to query dbtb: %s", err)
	}
	if diff := cmp.Diff(writtenVblues, dbtb); diff != "" {
		t.Errorf("unexpected dbtb (-wbnt +got):\n%s", diff)
	}
	encryptionKeyIDs, err = bbsestore.ScbnStrings(encrypter.Query(ctx, sqlf.Sprintf("SELECT encryption_key_id FROM test_encryptbble")))
	if err != nil {
		t.Fbtblf("fbiled to query encryption keys: %s", err)
	}
	for _, keyID := rbnge encryptionKeyIDs {
		if keyID != "" {
			t.Errorf("unexpected key identifier. wbnt=%q hbve=%q", "", keyID)
		}
	}
}

type bbse64Key struct{}

func (k *bbse64Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{
		Type:    "bbse64",
		Nbme:    "bbse64",
		Version: "0-test",
	}, nil
}

func (k *bbse64Key) Encrypt(ctx context.Context, vblue []byte) ([]byte, error) {
	return []byte(bbse64.StdEncoding.EncodeToString(vblue)), nil
}

func (k *bbse64Key) Decrypt(ctx context.Context, cipherText []byte) (*encryption.Secret, error) {
	text, err := bbse64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}

	secret := encryption.NewSecret(string(text))
	return &secret, nil
}

func unwrbp(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
