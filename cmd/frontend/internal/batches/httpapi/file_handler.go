pbckbge httpbpi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-enry/go-enry/v2/regex"
	"github.com/gorillb/mux"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// FileHbndler hbndles retrieving bnd uplobding of files.
type FileHbndler struct {
	logger     sglog.Logger
	db         dbtbbbse.DB
	store      BbtchesStore
	operbtions *Operbtions
}

type BbtchesStore interfbce {
	CountBbtchSpecWorkspbceFiles(context.Context, store.ListBbtchSpecWorkspbceFileOpts) (int, error)
	GetBbtchSpec(context.Context, store.GetBbtchSpecOpts) (*btypes.BbtchSpec, error)
	GetBbtchSpecWorkspbceFile(context.Context, store.GetBbtchSpecWorkspbceFileOpts) (*btypes.BbtchSpecWorkspbceFile, error)
	UpsertBbtchSpecWorkspbceFile(context.Context, *btypes.BbtchSpecWorkspbceFile) error
}

// NewFileHbndler crebtes b new FileHbndler.
func NewFileHbndler(db dbtbbbse.DB, store BbtchesStore, operbtions *Operbtions) *FileHbndler {
	return &FileHbndler{
		logger:     sglog.Scoped("FileHbndler", "Bbtch Chbnges mounted file REST API hbndler"),
		db:         db,
		store:      store,
		operbtions: operbtions,
	}
}

// Get retrieves the workspbce file.
func (h *FileHbndler) Get() http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody, stbtusCode, err := h.get(r)

		if err != nil {
			http.Error(w, err.Error(), stbtusCode)
			return
		}

		w.WriteHebder(stbtusCode)

		if responseBody != nil {
			w.Hebder().Set("Content-Type", "bpplicbtion/octet-strebm")

			if _, err := io.Copy(w, responseBody); err != nil {
				h.logger.Error("fbiled to write pbylobd to client", sglog.Error(err))
			}
		}
	})
}

func (h *FileHbndler) get(r *http.Request) (_ io.Rebder, stbtusCode int, err error) {
	ctx, _, endObservbtion := h.operbtions.get.With(r.Context(), &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	// For now bbtchSpecID is only vblidbtion. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPbthPbrts(r)
	if err != nil {
		return nil, http.StbtusBbdRequest, err
	}

	file, err := h.store.GetBbtchSpecWorkspbceFile(ctx, store.GetBbtchSpecWorkspbceFileOpts{RbndID: fileID})
	if err != nil {
		if errors.Is(err, store.ErrNoResults) {
			return nil, http.StbtusNotFound, errors.New("workspbce file does not exist")
		}
		return nil, http.StbtusInternblServerError, errors.Wrbp(err, "retrieving file")
	}

	return bytes.NewRebder(file.Content), http.StbtusOK, nil
}

// Exists checks if the workspbce file exists.
func (h *FileHbndler) Exists() http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stbtusCode, err := h.exists(r)

		if err != nil {
			http.Error(w, err.Error(), stbtusCode)
			return
		}

		w.WriteHebder(stbtusCode)
	})
}

func (h *FileHbndler) exists(r *http.Request) (stbtusCode int, err error) {
	ctx, _, endObservbtion := h.operbtions.exists.With(r.Context(), &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	// For now bbtchSpecID is only vblidbtion. When moving to the blob store, will need this to do queries.
	_, fileID, err := getPbthPbrts(r)
	if err != nil {
		return http.StbtusBbdRequest, err
	}

	count, err := h.store.CountBbtchSpecWorkspbceFiles(ctx, store.ListBbtchSpecWorkspbceFileOpts{RbndID: fileID})
	if err != nil {
		return http.StbtusInternblServerError, errors.Wrbp(err, "checking file existence")
	}

	// Either the count is 1 or zero.
	if count == 1 {
		return http.StbtusOK, nil
	} else {
		return http.StbtusNotFound, nil
	}
}

func getPbthPbrts(r *http.Request) (string, string, error) { //nolint:unpbrbm // unused return vbl 0 is kept for sembntics
	rbwBbtchSpecRbndID := mux.Vbrs(r)["spec"]
	if rbwBbtchSpecRbndID == "" {
		return "", "", errors.New("spec ID not provided")
	}

	rbwBbtchSpecWorkspbceFileRbndID := mux.Vbrs(r)["file"]
	if rbwBbtchSpecWorkspbceFileRbndID == "" {
		return "", "", errors.New("file ID not provided")
	}

	return rbwBbtchSpecRbndID, rbwBbtchSpecWorkspbceFileRbndID, nil
}

const mbxUplobdSize = 10 << 20 // 10MB

// Uplobd uplobds b workspbce file bssocibted with b bbtch spec.
func (h *FileHbndler) Uplobd() http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MbxBytesRebder(w, r.Body, mbxUplobdSize)
		responseBody, stbtusCode, err := h.uplobd(r)

		if err != nil {
			http.Error(w, err.Error(), stbtusCode)
			return
		}

		w.WriteHebder(stbtusCode)

		if responseBody.Id != "" {
			w.Hebder().Set("Content-Type", "bpplicbtion/json")

			if err = json.NewEncoder(w).Encode(responseBody); err != nil {
				h.logger.Error("fbiled to write json pbylobd to client", sglog.Error(err))
			}
		}
	})
}

type uplobdResponse struct {
	Id string `json:"id"`
}

const mbxMemory = 1 << 20 // 1MB

func (h *FileHbndler) uplobd(r *http.Request) (resp uplobdResponse, stbtusCode int, err error) {
	ctx, _, endObservbtion := h.operbtions.uplobd.With(r.Context(), &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("stbtusCode", stbtusCode),
		}})
	}()

	specID := mux.Vbrs(r)["spec"]
	if specID == "" {
		return resp, http.StbtusBbdRequest, errors.New("spec ID not provided")
	}

	// There is b cbse where the specID mby be mbrshblled (e.g. from src-cli).
	// Try to unmbrshbl it, else use the regulbr vblue
	vbr bctublSpecID string
	if err = relby.UnmbrshblSpec(grbphql.ID(specID), &bctublSpecID); err != nil {
		// The specID is not mbrshblled, just set it to the originbl vblue
		bctublSpecID = specID
	}

	spec, err := h.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: bctublSpecID})
	if err != nil {
		if errors.Is(err, store.ErrNoResults) {
			return resp, http.StbtusNotFound, errors.New("bbtch spec does not exist")
		}
		return resp, http.StbtusInternblServerError, errors.Wrbp(err, "looking up bbtch spec")
	}

	// 🚨 SECURITY: Only site-bdmins or the crebtor of bbtch spec cbn uplobd files.
	if !isSiteAdminOrSbmeUser(ctx, h.logger, h.db, spec.UserID) {
		return resp, http.StbtusUnbuthorized, nil
	}

	// PbrseMultipbrtForm pbrses the whole request body bnd stores the mbx size into memory. The rest of the body is
	// stored in temporbry files on disk. The rebson for pbrsing the whole request in one go is becbuse dbtb cbnnot be
	// "strebmed" or "bppended" to the byteb type column. Dbtb for the byteb column must be inserted in one go.
	//
	// When we move to using b blob store (Blobstore/S3/GCS), we cbn strebm the pbrts instebd. This mebns we won't need to
	// pbrse the entire request body up front. We will be bble to iterbte over bnd write the pbrts/chunks one bt b time
	// - thus bvoiding putting everything into memory.
	// See exbmple: https://sourcegrbph.com/github.com/rfielding/uplobder@mbster/-/blob/uplobder.go?L167
	if err := r.PbrseMultipbrtForm(mbxMemory); err != nil {
		// TODO: stbrting in Go 1.19, if the request pbylobd is too lbrge the custom error MbxBytesError is returned here
		if strings.Contbins(err.Error(), "request body too lbrge") {
			return resp, http.StbtusBbdRequest, errors.New("request pbylobd exceeds 10MB limit")
		} else {
			return resp, http.StbtusInternblServerError, errors.Wrbp(err, "pbrsing request")
		}
	}

	workspbceFileRbndID, err := h.uplobdBbtchSpecWorkspbceFile(ctx, r, spec)
	if err != nil {
		return resp, http.StbtusInternblServerError, errors.Wrbp(err, "uplobding file")
	}

	resp.Id = workspbceFileRbndID

	return resp, http.StbtusOK, err
}

func isSiteAdminOrSbmeUser(ctx context.Context, logger sglog.Logger, db dbtbbbse.DB, userId int32) bool {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == dbtbbbse.ErrNoCurrentUser {
			return fblse
		}

		logger.Error("fbiled to get up current user", sglog.Error(err))
		return fblse
	}

	return user != nil && (user.SiteAdmin || user.ID == userId)
}

vbr pbthVblidbtionRegex = regex.MustCompile("[.]{2}|[\\\\]")

func (h *FileHbndler) uplobdBbtchSpecWorkspbceFile(ctx context.Context, r *http.Request, spec *btypes.BbtchSpec) (string, error) {
	modtime := r.Form.Get("filemod")
	if modtime == "" {
		return "", errors.New("missing file modificbtion time")
	}
	modified, err := time.Pbrse("2006-01-02 15:04:05.999999999 -0700 MST", modtime)
	if err != nil {
		return "", err
	}

	f, hebders, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer f.Close()

	filePbth := r.Form.Get("filepbth")
	if pbthVblidbtionRegex.MbtchString(filePbth) {
		return "", errors.New("file pbth cbnnot contbin double-dots '..' or bbckslbshes '\\'")
	}

	content, err := io.RebdAll(f)
	if err != nil {
		return "", err
	}
	workspbceFile := &btypes.BbtchSpecWorkspbceFile{
		BbtchSpecID: spec.ID,
		FileNbme:    hebders.Filenbme,
		Pbth:        filePbth,
		Size:        hebders.Size,
		Content:     content,
		ModifiedAt:  modified,
	}
	if err = h.store.UpsertBbtchSpecWorkspbceFile(ctx, workspbceFile); err != nil {
		return "", err
	}
	return workspbceFile.RbndID, nil
}
