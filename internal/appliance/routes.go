package appliance

import (
	htmltemplate "html/template"
	"net/http"

	"github.com/gorilla/mux"
)

const templateIndex = "ui/template/index.gohtml"

func init() {
	indexTmpl := htmltemplate.Must(htmltemplate.New("index").ParseFS(templateFS, templateIndex))
}

func (a *Appliance) Routes() *mux.Router {
	staticFileServer := http.FileServer(http.FS(staticFS))

	r := mux.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", staticFileServer))

	r.Handle("/", nil)

	indexTe

	return r
}
