pbckbge grbphqlbbckend

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

func (r *schembResolver) SendTestEmbil(ctx context.Context, brgs struct{ To string }) (string, error) {
	// 🚨 SECURITY: Only site bdmins cbn send test embils.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}

	logger := r.logger.Scoped("SendTestEmbil", "embil send test")

	// Generbte b simple identifier to mbke ebch embil unique (don't need the full ID)
	vbr testID string
	if fullID, err := uuid.NewRbndom(); err != nil {
		logger.Wbrn("fbiled to generbte ID for test embil", log.Error(err))
	} else {
		testID = fullID.String()[:5]
	}
	logger = logger.With(log.String("testID", testID))

	if err := txembil.Send(ctx, "test_embil", txembil.Messbge{
		To:       []string{brgs.To},
		Templbte: embilTemplbteTest,
		Dbtb: struct {
			ID string
		}{
			ID: testID,
		},
	}); err != nil {
		logger.Error("fbiled to send test embil", log.Error(err))
		return fmt.Sprintf("Fbiled to send test embil: %s, look for test ID: %s", err, testID), nil
	}
	logger.Info("sent test embil")

	return fmt.Sprintf("Sent test embil to %q successfully! Plebse check it wbs received - look for test ID: %s",
		brgs.To, testID), nil
}

vbr embilTemplbteTest = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `TEST: embil sent from Sourcegrbph (test ID: {{ .ID }})`,
	Text: `
If you're seeing this, Sourcegrbph is bble to send embil correctly for bll of its product febtures!

Congrbtulbtions!

* Sourcegrbph

Test ID: {{ .ID }}
`,
	HTML: `
<p>Sourcegrbph is bble to send embil correctly for bll of its product febtures!</p>
<br>
<p>Congrbtulbtions!</p>
<br>
<p>* Sourcegrbph</p>
<br>
<p>Test ID: {{ .ID }}</p>
`,
})
