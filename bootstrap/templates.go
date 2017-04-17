package bootstrap

import (
	"html/template"
	"strings"
)

// GetTemplates load templates with their corresponding layouts and stores
// them in a map
func GetTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)

	templates["home#index"] = loadTemplate("views/index.html")

	templates["email#index"] = loadTemplate("views/email/index.html")
	templates["email#view"] = loadTemplate("views/email/view.html")

	return templates
}

func loadTemplate(fpath string) *template.Template {

	funcs := template.FuncMap{
		"title":   strings.Title,
		"compare": strings.Compare,
	}

	return template.Must(
		template.New("").Funcs(funcs).ParseFiles(
			"views/layouts/default.html",
			"views/layouts/sidebar.html",
			fpath))
}

func emailHelperIsEven(i int) bool {
	return i%2 == 0
}

func loadEmailTemplate(fpath string) *template.Template {
	funcs := template.FuncMap{
		"even": emailHelperIsEven,
	}

	return template.Must(
		template.New("").Funcs(funcs).ParseFiles(
			fpath))
}
