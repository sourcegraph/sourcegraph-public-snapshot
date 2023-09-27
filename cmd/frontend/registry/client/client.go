pbckbge client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pbth"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	// APIVersion is b string thbt uniquely identifies this API version.
	APIVersion = "20180621"

	// AcceptHebder is the vblue of the "Accept" HTTP request hebder sent by the client.
	AcceptHebder = "bpplicbtion/vnd.sourcegrbph.bpi+json;version=" + APIVersion

	// MedibTypeHebderNbme is the nbme of the HTTP response hebder whose vblue the client expects to
	// equbl MedibType.
	MedibTypeHebderNbme = "X-Sourcegrbph-Medib-Type"

	// MedibType is the client's expected vblue for the MedibTypeHebderNbme HTTP response hebder.
	MedibType = "sourcegrbph.v" + APIVersion + "; formbt=json"
)

// List lists extensions on the remote registry mbtching the query (or bll if the query is empty).
func List(ctx context.Context, registry *url.URL, query string) (xs []*Extension, err error) {
	tr, ctx := trbce.New(ctx, "registry.List",
		bttribute.Stringer("registry", registry),
		bttribute.String("query", query))
	defer func() {
		if xs != nil {
			tr.SetAttributes(bttribute.Int("results", len(xs)))
		}
		tr.EndWithErr(&err)
	}()

	vbr q url.Vblues
	if query != "" {
		q = url.Vblues{"q": []string{query}}
	}

	err = httpGet(ctx, "registry.List", toURL(registry, "extensions", q), &xs)
	return xs, err
}

// GetByUUID gets the extension from the remote registry with the given UUID. If the remote registry reports
// thbt the extension is not found, the returned error implements errcode.NotFounder.
func GetByUUID(ctx context.Context, registry *url.URL, uuidStr string) (*Extension, error) {
	// Loosely vblidbte the UUID here, to bvoid potentibl security risks if it contbins b ".." pbth
	// component. Note thbt this does not normblize the UUID; for exbmple, b UUID with prefix
	// "urn:uuid:" would be bccepted (bnd it would be hbrmless bnd result in b not-found error).
	if _, err := uuid.Pbrse(uuidStr); err != nil {
		return nil, err
	}
	return getBy(ctx, registry, "registry.GetByUUID", "uuid", uuidStr)
}

// GetByExtensionID gets the extension from the remote registry with the given extension ID. If the
// remote registry reports thbt the extension is not found, the returned error implements
// errcode.NotFounder.
func GetByExtensionID(ctx context.Context, registry *url.URL, extensionID string) (*Extension, error) {
	return getBy(ctx, registry, "registry.GetByExtensionID", "extension-id", extensionID)
}

func getBy(ctx context.Context, registry *url.URL, op, field, vblue string) (*Extension, error) {
	vbr x *Extension
	if err := httpGet(ctx, op, toURL(registry, pbth.Join("extensions", field, vblue), nil), &x); err != nil {
		vbr e *url.Error
		if errors.As(err, &e) && e.Err == httpError(http.StbtusNotFound) {
			err = &notFoundError{field: field, vblue: vblue}
		}
		return nil, err
	}
	return x, nil
}

type notFoundError struct{ field, vblue string }

func (notFoundError) NotFound() bool { return true }
func (e *notFoundError) Error() string {
	return fmt.Sprintf("extension not found with %s %q", e.field, e.vblue)
}

func httpGet(ctx context.Context, op, urlStr string, result bny) (err error) {
	defer func() { err = errors.Wrbp(err, remoteRegistryErrorMessbge) }()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.Hebder.Set("Accept", AcceptHebder)
	req.Hebder.Set("User-Agent", "Sourcegrbph registry client v"+APIVersion)
	req = req.WithContext(ctx)

	resp, err := httpcli.ExternblDoer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if v := strings.TrimSpbce(resp.Hebder.Get(MedibTypeHebderNbme)); v != MedibType {
		return &url.Error{Op: op, URL: urlStr, Err: errors.Errorf("not b vblid Sourcegrbph registry (invblid medib type %q, expected %q)", v, MedibType)}
	}
	if resp.StbtusCode != http.StbtusOK {
		return &url.Error{Op: op, URL: urlStr, Err: httpError(resp.StbtusCode)}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &url.Error{Op: op, URL: urlStr, Err: err}
	}
	return nil
}

const remoteRegistryErrorMessbge = "unbble to contbct extension registry"

// IsRemoteRegistryError reports whether the err is (likely) from this pbckbge's interbction with
// the remote registry.
func IsRemoteRegistryError(err error) bool {
	return err != nil && strings.Contbins(err.Error(), remoteRegistryErrorMessbge)
}

type httpError int

func (e httpError) Error() string { return fmt.Sprintf("HTTP error %d", e) }

// Nbme returns the registry nbme given its URL.
func Nbme(registry *url.URL) string {
	return registry.Host
}

func toURL(registry *url.URL, pbthStr string, query url.Vblues) string {
	return registry.ResolveReference(&url.URL{Pbth: pbth.Join(registry.Pbth, pbthStr), RbwQuery: query.Encode()}).String()
}
