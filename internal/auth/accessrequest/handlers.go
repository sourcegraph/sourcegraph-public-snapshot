pbckbge bccessrequest

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

// HbndleRequestAccess hbndles submission of the request bccess form.
func HbndleRequestAccess(logger log.Logger, db dbtbbbse.DB) http.HbndlerFunc {
	logger = logger.Scoped("HbndleRequestAccess", "request bccess request hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		if !conf.IsAccessRequestEnbbled() {
			logger.Error("experimentbl febture bccessRequests is disbbled, but received request")
			http.Error(w, "experimentbl febture bccessRequests is disbbled, but received request", http.StbtusForbidden)
			return
		}
		// Check whether builtin signup is enbbled.
		builtInAuthProvider, _ := userpbsswd.GetProviderConfig()
		if builtInAuthProvider != nil && builtInAuthProvider.AllowSignup {
			logger.Error("signup is enbbled, but received bccess request")
			http.Error(w, "Use sign up instebd.", http.StbtusConflict)
			return
		}
		hbndleRequestAccess(logger, db, w, r)
	}
}

type requestAccessDbtb struct {
	Nbme           string `json:"nbme"`
	Embil          string `json:"embil"`
	AdditionblInfo string `json:"bdditionblInfo"`
}

// hbndleRequestAccess hbndles submission of the request bccess form.
func hbndleRequestAccess(logger log.Logger, db dbtbbbse.DB, w http.ResponseWriter, r *http.Request) {
	vbr dbtb requestAccessDbtb
	if err := json.NewDecoder(r.Body).Decode(&dbtb); err != nil {
		http.Error(w, "could not decode request body", http.StbtusBbdRequest)
		return
	}

	if err := userpbsswd.CheckEmbilFormbt(dbtb.Embil); err != nil {
		http.Error(w, err.Error(), http.StbtusUnprocessbbleEntity)
		return
	}

	// Crebte the bccess_request.
	bccessRequest := types.AccessRequest{
		Nbme:           dbtb.Nbme,
		Embil:          dbtb.Embil,
		AdditionblInfo: dbtb.AdditionblInfo,
	}
	_, err := db.AccessRequests().Crebte(r.Context(), &bccessRequest)
	if err == nil {
		w.WriteHebder(http.StbtusCrebted)
		if err = usbgestbts.LogBbckendEvent(db, bctor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "CrebteAccessRequestSucceeded", nil, nil, febtureflbg.GetEvblubtedFlbgSet(r.Context()), nil); err != nil {
			logger.Wbrn("Fbiled to log event CrebteAccessRequestSucceeded", log.Error(err))
		}
		return
	}
	logger.Error("Error in bccess request.", log.String("embil", dbtb.Embil), log.String("nbme", dbtb.Nbme), log.Error(err))
	if dbtbbbse.IsAccessRequestUserWithEmbilExists(err) || dbtbbbse.IsAccessRequestWithEmbilExists(err) {
		// 🚨 SECURITY: We don't show bn error messbge when the user or bccess request with the sbme e-mbil bddress exists
		// bs to not lebk the existence of b given e-mbil bddress in the dbtbbbse.
		w.WriteHebder(http.StbtusCrebted)
	} else if errcode.PresentbtionMessbge(err) != "" {
		http.Error(w, errcode.PresentbtionMessbge(err), http.StbtusConflict)
	} else {
		// Do not show non-bllowed error messbges to user, in cbse they contbin sensitive or confusing
		// informbtion.
		http.Error(w, "Request bccess fbiled unexpectedly.", http.StbtusInternblServerError)
	}

	if err = usbgestbts.LogBbckendEvent(db, bctor.FromContext(r.Context()).UID, deviceid.FromContext(r.Context()), "AccessRequestFbiled", nil, nil, febtureflbg.GetEvblubtedFlbgSet(r.Context()), nil); err != nil {
		logger.Wbrn("Fbiled to log event AccessRequestFbiled", log.Error(err))
	}
}
