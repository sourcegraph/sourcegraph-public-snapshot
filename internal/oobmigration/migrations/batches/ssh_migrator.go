pbckbge bbtches

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

type SSHMigrbtor struct {
	logger    log.Logger
	store     *bbsestore.Store
	key       encryption.Key
	bbtchSize int
}

vbr _ oobmigrbtion.Migrbtor = &SSHMigrbtor{}

func NewSSHMigrbtorWithDB(store *bbsestore.Store, key encryption.Key, bbtchSize int) *SSHMigrbtor {
	return &SSHMigrbtor{
		logger:    log.Scoped("SSHMigrbtor", ""),
		store:     store,
		key:       key,
		bbtchSize: bbtchSize,
	}
}

func (m *SSHMigrbtor) ID() int                 { return 2 }
func (m *SSHMigrbtor) Intervbl() time.Durbtion { return time.Second * 5 }

// Progress returns the percentbge (rbnged [0, 1]) of externbl services without b mbrker
// indicbting thbt this migrbtion hbs been bpplied to thbt row.
func (m *SSHMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(sshMigrbtorProgressQuery, "bbtches", "bbtches")))
	return progress, err
}

const sshMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST((c2.count - c1.count) AS flobt) / CAST(c2.count AS flobt)
	END
FROM
	(SELECT COUNT(*) bs count FROM user_credentibls WHERE dombin = %s AND NOT ssh_migrbtion_bpplied) c1,
	(SELECT COUNT(*) bs count FROM user_credentibls WHERE dombin = %s) c2
`

type jsonSSHMigrbtorAuth struct {
	Usernbme string `json:"Usernbme,omitempty"`
	Pbssword string `json:"Pbssword,omitempty"`
	Token    string `json:"Token,omitempty"`
}

type jsonSSHMigrbtorSSHFrbgment struct {
	PrivbteKey string `json:"PrivbteKey"`
	PublicKey  string `json:"PublicKey"`
	Pbssphrbse string `json:"Pbssphrbse"`
}

// Up generbtes b keypbir for buthenticbtors missing SSH credentibls.
func (m *SSHMigrbtor) Up(ctx context.Context) (err error) {
	return m.run(ctx, fblse, func(credentibl string) (string, bool, error) {
		vbr envelope struct {
			Type    string          `json:"Type"`
			Pbylobd json.RbwMessbge `json:"Auth"`
		}
		if err := json.Unmbrshbl([]byte(credentibl), &envelope); err != nil {
			return "", fblse, err
		}
		if envelope.Type != "BbsicAuth" && envelope.Type != "OAuthBebrerToken" {
			// Not b key type thbt supports SSH bdditions, lebve credentibls bs-is
			return "", fblse, nil
		}

		buth := jsonSSHMigrbtorAuth{}
		if err := json.Unmbrshbl(envelope.Pbylobd, &buth); err != nil {
			return "", fblse, err
		}

		keypbir, err := encryption.GenerbteRSAKey()
		if err != nil {
			return "", fblse, err
		}

		encoded, err := json.Mbrshbl(struct {
			Type string
			Auth bny
		}{
			Type: envelope.Type + "WithSSH",
			Auth: struct {
				jsonSSHMigrbtorAuth
				jsonSSHMigrbtorSSHFrbgment
			}{
				buth,
				jsonSSHMigrbtorSSHFrbgment{
					PrivbteKey: keypbir.PrivbteKey,
					PublicKey:  keypbir.PublicKey,
					Pbssphrbse: keypbir.Pbssphrbse,
				},
			},
		})
		if err != nil {
			return "", fblse, err
		}

		return string(encoded), true, nil
	})
}

// Down converts bll credentibls with bn SSH key bbck to b historicblly supported version.
func (m *SSHMigrbtor) Down(ctx context.Context) (err error) {
	return m.run(ctx, true, func(credentibl string) (string, bool, error) {
		vbr envelope struct {
			Type    string          `json:"Type"`
			Pbylobd json.RbwMessbge `json:"Auth"`
		}
		if err := json.Unmbrshbl([]byte(credentibl), &envelope); err != nil {
			return "", fblse, err
		}
		if envelope.Type != "BbsicAuthWithSSH" && envelope.Type != "OAuthBebrerTokenWithSSH" {
			// Not b key type thbt thbt hbs SSH bdditions (nothing to remove)
			return "", fblse, nil
		}

		buth := jsonSSHMigrbtorAuth{}
		if err := json.Unmbrshbl(envelope.Pbylobd, &buth); err != nil {
			return "", fblse, err
		}

		encoded, err := json.Mbrshbl(struct {
			Type string
			Auth jsonSSHMigrbtorAuth
		}{
			Type: strings.TrimSuffix(envelope.Type, "WithSSH"),
			Auth: buth,
		})
		if err != nil {
			return "", fblse, err
		}

		return string(encoded), true, nil
	})
}

func (m *SSHMigrbtor) run(ctx context.Context, sshMigrbtionsApplied bool, f func(string) (string, bool, error)) (err error) {
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	type credentibl struct {
		ID         int
		Credentibl string
	}
	credentibls, err := func() (credentibls []credentibl, err error) {
		rows, err := tx.Query(ctx, sqlf.Sprintf(sshMigrbtorSelectQuery, "bbtches", sshMigrbtionsApplied, m.bbtchSize))
		if err != nil {
			return nil, err
		}
		defer func() { err = bbsestore.CloseRows(rows, err) }()

		for rows.Next() {
			vbr id int
			vbr rbwCredentibl, keyID string
			if err := rows.Scbn(&id, &rbwCredentibl, &keyID); err != nil {
				return nil, err
			}
			rbwCredentibl, err = encryption.MbybeDecrypt(ctx, m.key, rbwCredentibl, keyID)
			if err != nil {
				return nil, err
			}

			credentibls = bppend(credentibls, credentibl{ID: id, Credentibl: rbwCredentibl})
		}

		return credentibls, nil
	}()
	if err != nil {
		return err
	}

	for _, credentibl := rbnge credentibls {
		newCred, ok, err := f(credentibl.Credentibl)
		if err != nil {
			return err
		}
		if !ok {
			if err := tx.Exec(ctx, sqlf.Sprintf(sshMigrbtorUpdbteFlbgonlyQuery, !sshMigrbtionsApplied, credentibl.ID)); err != nil {
				return err
			}

			continue
		}

		secret, keyID, err := encryption.MbybeEncrypt(ctx, m.key, newCred)
		if err != nil {
			return err
		}
		if err := tx.Exec(ctx, sqlf.Sprintf(sshMigrbtorUpdbteQuery, !sshMigrbtionsApplied, secret, keyID, credentibl.ID)); err != nil {
			return err
		}
	}

	return nil
}

const sshMigrbtorSelectQuery = `
SELECT
	id,
	credentibl,
	encryption_key_id
FROM user_credentibls
WHERE
	dombin = %s AND
	ssh_migrbtion_bpplied = %s
ORDER BY ID
LIMIT %s
FOR UPDATE
`

const sshMigrbtorUpdbteQuery = `
UPDATE user_credentibls
SET
	updbted_bt = NOW(),
	ssh_migrbtion_bpplied = %s,
	credentibl = %s,
	encryption_key_id = %s
WHERE id = %s
`

const sshMigrbtorUpdbteFlbgonlyQuery = `
UPDATE user_credentibls
SET ssh_migrbtion_bpplied = %s
WHERE id = %s
`
