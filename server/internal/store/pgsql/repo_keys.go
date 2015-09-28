package pgsql

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/store"
)

// repoKey is a private key used to access a remote repository origin.
type repoKey struct {
	Repo          string // URI of repository this key accesses.
	PrivateKeyPEM string `db:"private_key_pem"`
}

func init() {
	tbl := Schema.Map.AddTableWithName(repoKey{}, "repo_key").SetKeys(false, "Repo")
	tbl.ColMap("PrivateKeyPEM").SetSqlType("text")
}

// MirroredRepoSSHKeys is a DB-backed implementation of the MirroredRepoSSHKeys store.
type MirroredRepoSSHKeys struct{}

var _ store.MirroredRepoSSHKeys = (*MirroredRepoSSHKeys)(nil)

func (s *MirroredRepoSSHKeys) Create(ctx context.Context, repo string, privKey *rsa.PrivateKey) error {
	block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privKey), []byte(pemPassword), x509.PEMCipherAES128)
	if err != nil {
		return err
	}
	pemBytes := pem.EncodeToMemory(block)

	sql := `
WITH update_result AS (
  UPDATE repo_key SET private_key_pem=$2 WHERE repo=$1
  RETURNING 1
),
insert_data AS (
  SELECT $1 AS repo, $2 AS private_key_pem
)
INSERT INTO repo_key(repo, private_key_pem)
SELECT * FROM insert_data
WHERE NOT EXISTS (SELECT NULL FROM update_result);
`
	if _, err := dbh(ctx).Exec(sql, repo, string(pemBytes)); err != nil {
		return err
	}
	return nil
}

func (s *MirroredRepoSSHKeys) GetPEM(ctx context.Context, repo string) ([]byte, error) {
	var k []*repoKey
	if err := dbh(ctx).Select(&k, `SELECT * FROM repo_key WHERE repo=$1 LIMIT 1`, repo); err != nil {
		return nil, err
	}
	if len(k) == 0 {
		return nil, nil
	}

	block, _ := pem.Decode([]byte(k[0].PrivateKeyPEM))
	d, err := x509.DecryptPEMBlock(block, []byte(pemPassword))
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: d}), nil
}

func (s *MirroredRepoSSHKeys) Delete(ctx context.Context, repo string) error {
	_, err := dbh(ctx).Exec(`DELETE FROM repo_key WHERE repo=$1;`, repo)
	return err
}

// pemPassword is the passphrase used to encrypt and decrypt private
// keys stored in the DB.
var pemPassword = os.Getenv("SG_PEM_ENCRYPTION_PASSWORD")

func init() {
	if pemPassword == "" && conf.RequireSecrets {
		log.Fatalf("SG_PEM_ENCRYPTION_PASSWORD env var must not be empty.")
	}
}
