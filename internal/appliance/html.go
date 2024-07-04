package appliance

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/life4/genesis/slices"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	formValueOn = "on"
)

func templatePath(name string) string {
	return filepath.Join("web", "template", name+".gohtml")
}

func (a *Appliance) applianceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok, _ := a.shouldSetupRun(context.Background()); ok {
			http.Redirect(w, r, "/appliance/setup", http.StatusSeeOther)
		}
	})
}

func renderTemplate(name string, w io.Writer, data any) error {
	tmpl, err := template.ParseFS(templateFS, templatePath("layout"), templatePath(name))
	if err != nil {
		return errors.Wrapf(err, "rendering template: %s", name)
	}
	return tmpl.Execute(w, data)
}

func (a *Appliance) getSetupHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versions, err := a.getVersions(r.Context())
		if err != nil {
			a.handleError(w, err, "getting versions")
			return
		}
		versions, err = NMinorVersions(versions, a.latestSupportedVersion, 2)
		if err != nil {
			a.handleError(w, err, "filtering versions to 2 minor points")
			return
		}

		err = renderTemplate("setup", w, struct {
			Versions []string
		}{
			Versions: versions,
		})
		if err != nil {
			a.handleError(w, err, "executing template")
			return
		}
	})
}

func (a *Appliance) getLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(a.adminPasswordBcrypt) == 0 {
			msg := fmt.Sprintf(
				"You must set a password: please create a secret named '%s' with key '%s'.",
				initialPasswordSecretName,
				initialPasswordSecretPasswordKey,
			)
			a.redirectToErrorPage(w, r, msg, errors.New("no admin password set"), true)
			return
		}

		if err := renderTemplate("landing", w, struct {
			Flash string
		}{
			Flash: r.URL.Query().Get(queryKeyUserMessage),
		}); err != nil {
			a.handleError(w, err, "executing template")
			return
		}
	})
}

func (a *Appliance) postLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSuppliedPassword := r.FormValue("password")
		if err := bcrypt.CompareHashAndPassword(a.adminPasswordBcrypt, []byte(userSuppliedPassword)); err != nil {
			if err == bcrypt.ErrMismatchedHashAndPassword {
				a.redirectWithError(w, r, r.URL.Path, "Supplied password is incorrect.", err, true)
				return
			}

			a.redirectToErrorPage(w, r, errMsgSomethingWentWrong, err, false)
			return
		}

		if err := passwordvalidator.Validate(userSuppliedPassword, 60); err != nil {
			msg := fmt.Sprintf(
				"Please set a stronger password: delete the '%s' secret, and create a new secret named '%s' with key '%s'.",
				dataSecretName,
				initialPasswordSecretName,
				initialPasswordSecretPasswordKey,
			)
			a.redirectToErrorPage(w, r, msg, err, true)
			return
		}

		validUntil := time.Now().Add(time.Hour).UTC()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			jwtClaimsValidUntilKey: validUntil.Format(time.RFC3339),
		})
		tokenStr, err := token.SignedString(a.jwtSecret)
		if err != nil {
			a.handleError(w, err, errMsgSomethingWentWrong)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    authCookieName,
			Value:   tokenStr,
			Expires: validUntil,
		})
		http.Redirect(w, r, "/appliance", http.StatusFound)
	})
}

func (a *Appliance) handleError(w http.ResponseWriter, err error, msg string) {
	a.logger.Error(msg, log.Error(err))

	// TODO we should probably look twice at this and decide whether it's in
	// line with existing standards.
	// Don't leak details of internal errors to users - that's why we have
	// logging above.
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, errMsgSomethingWentWrong)
}

func (a *Appliance) getVersions(ctx context.Context) ([]string, error) {
	versions, err := a.releaseRegistryClient.ListVersions(ctx, "sourcegraph")
	if err != nil {
		return nil, err
	}
	return slices.MapFilter(versions, func(version releaseregistry.ReleaseInfo) (string, bool) {
		return version.Version, version.Public
	}), nil
}

func (a *Appliance) postSetupHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			a.logger.Error("failed to parse http form request", log.Error(err))
			// Handle err
		}

		a.sourcegraph.Spec.RequestedVersion = r.FormValue("version")
		if r.FormValue("external_database") == formValueOn {
			a.sourcegraph.Spec.PGSQL.DatabaseConnection = &config.DatabaseConnectionSpec{
				Host:     r.FormValue("pgsqlDBHost"),
				Port:     r.FormValue("pgsqlDBPort"),
				User:     r.FormValue("pgsqlDBUser"),
				Password: r.FormValue("pgsqlDBPassword"),
				Database: r.FormValue("pgsqlDBName"),
			}
			a.sourcegraph.Spec.CodeIntel.DatabaseConnection = &config.DatabaseConnectionSpec{
				Host:     r.FormValue("codeintelDBHost"),
				Port:     r.FormValue("codeintelDBPort"),
				User:     r.FormValue("codeintelDBUser"),
				Password: r.FormValue("codeintelDBPassword"),
				Database: r.FormValue("codeintelDBName"),
			}
			a.sourcegraph.Spec.CodeInsights.DatabaseConnection = &config.DatabaseConnectionSpec{
				Host:     r.FormValue("codeinsightsDBHost"),
				Port:     r.FormValue("codeinsightsDBPort"),
				User:     r.FormValue("codeinsightsDBUser"),
				Password: r.FormValue("codeinsightsDBPassword"),
				Database: r.FormValue("codeinsightsDBName"),
			}
		}
		// TODO validate user input

		if r.FormValue("dev_mode") == formValueOn {
			a.sourcegraph.SetLocalDevMode()
		}

		_, err = a.CreateConfigMap(r.Context(), "sourcegraph-appliance")
		if err != nil {
			a.logger.Error("failed to create configMap sourcegraph-appliance", log.Error(err))
			// Handle err
		}
		a.status = StatusInstalling

		http.Redirect(w, r, "/appliance", http.StatusSeeOther)
	})
}
