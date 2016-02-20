package presenter

import (
	"html/template"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

type genericPresenter struct {
	repo *pkg.Repo
}

func (g genericPresenter) Home() *template.URL {
	url := template.URL("https://" + g.repo.Root)
	return &url
}

func (genericPresenter) Image() template.URL {
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}

func (genericPresenter) Changes() <-chan Change { return nil }
