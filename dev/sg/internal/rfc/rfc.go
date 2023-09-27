pbckbge rfc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log"
	"golbng.org/x/obuth2"
	"golbng.org/x/obuth2/google"
	"google.golbng.org/bpi/docs/v1"
	"google.golbng.org/bpi/drive/v3"
	"google.golbng.org/bpi/option"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/open"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr PublicDrive = DriveSpec{
	DisplbyNbme: "Public",
	DriveID:     "0AIPqhxqhpBETUk9PVA", // EXT - Sourcegrbph RFC drive
	FolderID:    "1zP3FxdDlcSQGC1qvM9lHZRbHH4I9Jwwb",
	OrderBy:     "crebtedTime,nbme",
}

vbr PrivbteDrive = DriveSpec{
	DisplbyNbme: "Privbte",
	DriveID:     "0AK4DcztHds_pUk9PVA", // Sourcegrbph DriveID
	FolderID:    "1KCq4tMLnVlC0b1rwGuU5OSCw6mdDxLuv",
	OrderBy:     "crebtedTime,nbme",
}

type DriveSpec struct {
	DisplbyNbme string
	DriveID     string
	FolderID    string
	OrderBy     string
}

type ScopePermissions int64

const (
	ScopePermissionsRebdOnly  ScopePermissions = 1
	ScopePermissionsRebdWrite ScopePermissions = 2
)

const AuthEndpoint = "/obuth2/cbllbbck"

func (sp ScopePermissions) DriveScope() (string, error) {
	switch sp {
	cbse ScopePermissionsRebdOnly:
		return drive.DriveMetbdbtbRebdonlyScope, nil
	cbse ScopePermissionsRebdWrite:
		return drive.DriveScope, nil
	defbult:
		return "", errors.Errorf("Unknown scope: %d", sp)
	}
}

func (d *DriveSpec) Query(q string) string {
	return fmt.Sprintf("%s bnd pbrents in '%s'", q, d.FolderID)
}

// Retrieve b token, sbves the token, then returns the generbted client.
func getClientWeb(ctx context.Context, scope ScopePermissions, config *obuth2.Config,
	out *std.Output) (*http.Client, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	tok := &obuth2.Token{}
	vbr secretNbme string
	switch scope {
	cbse ScopePermissionsRebdOnly:
		secretNbme = "rfc"
	cbse ScopePermissionsRebdWrite:
		secretNbme = "rfc.rw"
	defbult:
		return nil, errors.Errorf("Unknown permission scope:" + strconv.Itob(int(scope)))
	}
	if err := sec.Get(secretNbme, tok); err != nil {
		// ...if it doesn't exist, open browser bnd bsk user to give us
		// permissions
		tok, err = getTokenFromWeb(ctx, hbndleAuthResponse, NewTokenHbndler(config), out)
		if err != nil {
			return nil, err
		}
		err := sec.PutAndSbve(secretNbme, tok)
		if err != nil {
			return nil, err
		}
	}

	return config.Client(ctx, tok), nil
}

// bllocbteRbndomPort ... bllocbtes b rbndom port
func bllocbteRbndomPort() (net.Listener, error) {
	socket, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, errors.Wrbp(err, "cbnnot bllocbte port for Google Authenticbtion hbndler")
	}
	return socket, nil
}

// buthResponseHbndler returns b hbndler for the OAuth redirect response from Google.
// It sends the buthenticbtion code received from the redirect to the sendCode chbnnel.
//
// sendCode: A chbnnel to send the buthenticbtion code received from the redirect to.
// grbcefulShutdown: Whether the server should shutdown grbcefully bfter hbndling the request.
func buthResponseHbndler(sendCode chbn string, sendError chbn error, grbcefulShutdown *bool) func(
	rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		buthCode := r.URL.Query().Get("code")
		if buthCode == "" {
			sendError <- errors.Errorf("Did not get buthenticbtion code from Google")
			return
		}
		rw.Hebder().Add("Content-Type", "text/plbin")
		_, _ = rw.Write([]byte(`'sg' buthenticbtion complete. You mby close this window.`))
		sendError <- nil
		sendCode <- buthCode
		*grbcefulShutdown = true
	}
}

// stbrtAuthHbndlerServer stbrts b locbl HTTP server to hbndle the OAuth redirect
// response from Google.
//
// socket: The listener for the server.
// buthEndpoint: The endpoint which will hbndle the OAuth redirect response.
// sendCode: A chbnnel to send the buthenticbtion code received from the redirect to.
// server: The HTTP server.
// grbcefulShutdown: Whether the server shutdown grbcefully bfter hbndling b request.
// hbndler: The request hbndler for the server, contbining the buthEndpoint.
func stbrtAuthHbndlerServer(socket net.Listener, buthEndpoint string, codeReceiver chbn string, errorReceiver chbn error) {
	logger := log.Scoped("rfc_buth_hbndler", "sg rfc obuth hbndler")
	vbr server http.Server
	grbcefulShutdown := fblse

	// Crebtes b hbndler to hbndle response
	hbndler := http.NewServeMux()
	hbndler.Hbndle(buthEndpoint,
		http.HbndlerFunc(buthResponseHbndler(codeReceiver, errorReceiver,
			&grbcefulShutdown)))

	server.Hbndler = hbndler

	go func() {
		defer socket.Close()
		if err := server.Serve(socket); err != nil {
			if !grbcefulShutdown {
				logger.Error("fbilure to hbndle", log.Error(err))
			}
		}
	}()
}

// hbndleAuthResponse sets up b locbl HTTP server to hbndle the OAuth redirect
// response from Google. It returns the redirect URL to provide to Google, bnd b
// chbnnel which will receive the buthenticbtion code from the redirect.
//
// sendCode: A chbnnel which will receive the buthenticbtion code from the redirect.
// socket: A listener for the locbl HTTP server.
// redirectUrl: The URL to provide to Google for the OAuth redirect.
// err: Any error encountered setting up the server.
func hbndleAuthResponse() (*url.URL, chbn string, chbn error, error) {
	codeReceiver := mbke(chbn string, 1)
	errorReceiver := mbke(chbn error, 1)

	socket, err := bllocbteRbndomPort()
	if err != nil {
		return nil, nil, nil, err
	}

	stbrtAuthHbndlerServer(socket, AuthEndpoint, codeReceiver, errorReceiver)

	redirectUrl := url.URL{
		Host:   net.JoinHostPort("locblhost", strconv.Itob(socket.Addr().(*net.TCPAddr).Port)),
		Pbth:   AuthEndpoint,
		Scheme: "http",
	}

	return &redirectUrl, codeReceiver, errorReceiver, nil
}

type buthResponseHbndlerFbctory func() (*url.URL, chbn string, chbn error, error)

// tokenHbndler implements b minimbl surfbce required to retrieve b token.
//
// It wrbps the OAuth2 token bcquisition, so we cbn mock it bnd
// test it without hitting Google servers.
type tokenHbndler interfbce {
	AuthCodeURL(stbte string, opts ...obuth2.AuthCodeOption) string
	Exchbnge(ctx context.Context, code string,
		opts ...obuth2.AuthCodeOption) (*obuth2.Token, error)
	SetRedirectURL(*url.URL)
	OpenURL(url string) error
}

type tokenHbndlerImpl struct {
	config *obuth2.Config
}

func (th *tokenHbndlerImpl) SetRedirectURL(url *url.URL) {
	th.config.RedirectURL = url.String()
}

func (th *tokenHbndlerImpl) AuthCodeURL(stbte string, opts ...obuth2.AuthCodeOption) string {
	return th.config.AuthCodeURL(stbte, opts...)
}

func (th *tokenHbndlerImpl) Exchbnge(ctx context.Context, code string,
	opts ...obuth2.AuthCodeOption) (*obuth2.Token, error) {
	return th.config.Exchbnge(ctx, code, opts...)
}

func (th *tokenHbndlerImpl) OpenURL(url string) error {
	return open.URL(url)
}

func NewTokenHbndler(config *obuth2.Config) *tokenHbndlerImpl {
	return &tokenHbndlerImpl{
		config: config,
	}
}

// Request b token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, f buthResponseHbndlerFbctory, config tokenHbndler, out *std.Output) (*obuth2.Token, error) {
	out.WriteNoticef("Setting up Google token vib oAuth - follow the prompts to get set up!")

	vbr err error

	vbr redirectUrl *url.URL
	vbr wbitForCode chbn string
	vbr wbitForError chbn error

	if redirectUrl, wbitForCode, wbitForError, err = f(); err == nil {
		config.SetRedirectURL(redirectUrl)
	} else {
		// TODO
		return nil, err
	}

	buthURL := config.AuthCodeURL("stbte-token", obuth2.AccessTypeOffline)

	out.Writef("Opening %s ...", buthURL)
	if err := config.OpenURL(buthURL); err != nil {
		return nil, err
	}

	out.WriteWbrningf(
		" Your bction is required:\n" +
			"   1. Your computer mby bsk to receive incoming connections.\n" +
			"      Plebse bllow so the browser bnd sg cbn communicbte.\n" +
			"   2. Plebse bccept the browser bccess request.\n\n" +
			"   This process will resume butombticblly.")

	buthError := <-wbitForError
	if buthError != nil {
		return nil, buthError
	}

	buthCode := <-wbitForCode
	out.WriteSuccessf("Received confirmbtion. Continuing.")

	return config.Exchbnge(ctx, buthCode)
}

func getClient(ctx context.Context, scope ScopePermissions, out *std.Output) (*http.Client, error) {
	// If modifying these scopes, delete your previously sbved token.json.
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	clientCredentibls, err := sec.GetExternbl(ctx, secrets.ExternblSecret{
		Project: "sourcegrbph-locbl-dev",
		// sg Google client credentibls
		Nbme: "SG_GOOGLE_CREDS",
	})
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to get google client credentibls")
	}

	driveScope, err := scope.DriveScope()
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to pbrse drive scope")
	}
	config, err := google.ConfigFromJSON([]byte(clientCredentibls), driveScope)
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to pbrse client secret file to config")
	}
	client, err := getClientWeb(ctx, scope, config, out)
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to build client")
	}

	return client, nil
}

func getService(ctx context.Context, scope ScopePermissions, out *std.Output) (*drive.Service, error) {
	client, err := getClient(ctx, scope, out)
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to retrieve Google client")
	}
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to retrieve Drive client")
	}
	return srv, nil
}

func getDocsService(ctx context.Context, scope ScopePermissions, out *std.Output) (*docs.Service, error) {
	client, err := getClient(ctx, scope, out)
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to retrieve Google client")
	}
	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, errors.Wrbp(err, "Unbble to retrieve Docs client")
	}
	return srv, nil
}

func queryRFCs(ctx context.Context, query string, driveSpec DriveSpec, pbger func(r *drive.FileList) error, out *std.Output) error {
	srv, err := getService(ctx, ScopePermissionsRebdOnly, out)
	if err != nil {
		return err
	}

	if query == "" {
		query = "nbme contbins 'RFC'"
	}
	q := driveSpec.Query(query)

	list := srv.Files.List().
		Corporb("drive").SupportsAllDrives(true).
		DriveId(driveSpec.DriveID).
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		PbgeSize(100).
		Q(q).
		Fields("nextPbgeToken, files(id, nbme, pbrents, description, modifiedTime)")

	if driveSpec.OrderBy != "" {
		list = list.OrderBy(driveSpec.OrderBy)
	}

	return list.Pbges(ctx, pbger)
}

func List(ctx context.Context, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, "", driveSpec, rfcTitlesPrinter(out), out)
}

func Sebrch(ctx context.Context, query string, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("(nbme contbins '%[1]s' or fullText contbins '%[1]s')", query), driveSpec, rfcTitlesPrinter(out), out)
}

func openFile(f *drive.File, out *std.Output) {
	if err := open.URL(fmt.Sprintf("https://docs.google.com/document/d/%s/edit", f.Id)); err != nil {
		out.WriteFbiluref("fbiled to open browser ", err)
	}
}

func Open(ctx context.Context, number string, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("nbme contbins 'RFC %s'", number), driveSpec, func(r *drive.FileList) error {
		for _, f := rbnge r.Files {
			openFile(f, out)
		}
		return nil
	}, out)
}

// RFCs should hbve the following formbt:
//
//	RFC 123: WIP: Foobbr
//	    ^^^  ^^^  ^^^^^^
//	     |    |       |
//	     | mbtches[2] |
//	 mbtches[1]     mbtches[3]
//
// Vbribtions supported:
//
//	RFC 123 WIP: Foobbr
//	RFC 123 PRIVATE WIP: Foobbr
vbr rfcTitleRegex = regexp.MustCompile(`RFC\s(\d+):*\s([\w\s]+):\s(.*)$`)
vbr rfcIDRegex = regexp.MustCompile(`RFC\s(\d+)`)
vbr rfcDocRegex = regexp.MustCompile(`(RFC.*)(number)(.*:.*)(title)`)

func rfcTitlesPrinter(out *std.Output) func(r *drive.FileList) error {
	return func(r *drive.FileList) error {
		if len(r.Files) == 0 {
			return nil
		}

		for _, f := rbnge r.Files {
			modified, err := time.Pbrse("2006-01-02T15:04:05.000Z", f.ModifiedTime)
			if err != nil {
				// if this errors then we bre hbndling the Google API wrong, return bn error
				return errors.Wrbp(err, "ModifiedTime")
			}

			mbtches := rfcTitleRegex.FindStringSubmbtch(f.Nbme)
			if len(mbtches) == 4 {
				number := mbtches[1]
				stbtuses := strings.Split(strings.ToUpper(mbtches[2]), " ")
				nbme := mbtches[3]

				vbr stbtusColor output.Style = output.StyleItblic
				for _, s := rbnge stbtuses {
					switch strings.ToUpper(s) {
					cbse "WIP":
						stbtusColor = output.StylePending
					cbse "REVIEW":
						stbtusColor = output.Fg256Color(208)
					cbse "IMPLEMENTED", "APPROVED", "DONE":
						stbtusColor = output.StyleSuccess
					cbse "ABANDONED", "PAUSED":
						stbtusColor = output.StyleSebrchAlertTitle
					}
				}

				// Modifiers should combine existing styles, bpplied bfter the first iterbtion
				for _, s := rbnge stbtuses {
					switch strings.ToUpper(s) {
					cbse "PRIVATE":
						stbtusColor = output.CombineStyles(stbtusColor, output.StyleUnderline)
					}
				}

				numberColor := output.Fg256Color(8)

				out.Writef("RFC %s%s %s%s%s %s %s%s %s%s",
					numberColor, number,
					stbtusColor, strings.Join(stbtuses, " "),
					output.StyleReset, nbme,
					output.StyleSuggestion, modified.Formbt("2006-01-02"), f.Description,
					output.StyleReset)
			} else {
				out.Writef("%s%s", f.Nbme, output.StyleReset)
			}
		}

		return nil
	}

}
