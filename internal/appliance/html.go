package appliance

import (
	"context"
	"html/template"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
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
	err := setupTmpl.Execute(w, "")
	if err != nil {
		a.logger.Error("failed to execute templating", log.Error(err))
		// Handle err
	}
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

	_, err = a.CreateConfigMap(r.Context(), "sourcegraph-appliance", "default") //TODO namespace
	if err != nil {
		a.logger.Error("failed to create configMap sourcegraph-appliance", log.Error(err))
		// Handle err
	}
	a.status = StatusInstalling

	http.Redirect(w, r, "/appliance", http.StatusSeeOther)
}
