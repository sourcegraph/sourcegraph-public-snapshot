package appliance

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/life4/genesis/slices"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
)

const (
	templateSetup = "web/template/setup.gohtml"
)

var (
	setupTmpl *template.Template
)

func init() {
	setupTmpl = template.Must(template.ParseFS(templateFS, templateSetup))
}

func (a *Appliance) applianceHandler(w http.ResponseWriter, r *http.Request) {
	if ok, _ := a.shouldSetupRun(context.Background()); ok {
		http.Redirect(w, r, "/appliance/setup", http.StatusSeeOther)
	}
}

func (a *Appliance) getSetupHandler(w http.ResponseWriter, r *http.Request) {
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

	err = setupTmpl.Execute(w, struct {
		Versions []string
	}{
		Versions: versions,
	})
	if err != nil {
		a.handleError(w, err, "executing template")
		return
	}
}

func (a *Appliance) handleError(w http.ResponseWriter, err error, msg string) {
	a.logger.Error(msg, log.Error(err))

	// TODO we should probably look twice at this and decide whether it's in
	// line with existing standards.
	// Don't leak details of internal errors to users - that's why we have
	// logging above.
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Something went wrong - please contact support.")
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

func (a *Appliance) postSetupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.logger.Error("failed to parse http form request", log.Error(err))
		// Handle err
	}

	a.sourcegraph.Spec.RequestedVersion = r.FormValue("version")
	if r.FormValue("external_database") == "yes" {
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

	_, err = a.CreateConfigMap(r.Context(), "sourcegraph-appliance")
	if err != nil {
		a.logger.Error("failed to create configMap sourcegraph-appliance", log.Error(err))
		// Handle err
	}
	a.status = StatusInstalling

	http.Redirect(w, r, "/appliance", http.StatusSeeOther)
}
