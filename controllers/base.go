package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/anishmgoyal/calagora-admin/services"
)

// ViewData encapsulates any information needed to render a view
type ViewData struct {
	Session *models.Session
	Data    interface{}
}

// Base contains all variables shared by controllers
var Base struct {
	// Templates contains all views available to controllers
	Templates map[string]*template.Template
	// DB is the database connection to be used by controllers and passed to models
	Db *sql.DB
}

// BaseInitialization initializes all controllers
func BaseInitialization(templates map[string]*template.Template, db *sql.DB) {
	Base.Templates = templates
	Base.Db = db
	emailInit()
}

// BaseViewData populates view data based on the request and response writer
func BaseViewData(w http.ResponseWriter, r *http.Request) ViewData {
	return ViewData{
		Session: services.GetSession(r),
	}
}

// RenderPlainView attempts to render a view with only the base view data
func RenderPlainView(w http.ResponseWriter, r *http.Request,
	templateName string) {

	RenderView(w, templateName, BaseViewData(w, r))
}

// RenderView attempts to render a view. Gives a 404 error on failure
func RenderView(w http.ResponseWriter, templateName string, data ViewData) {
	var buff bytes.Buffer

	tmpl, ok := Base.Templates[templateName]
	if !ok {
		errStr := fmt.Sprintf("Unknown Template: %s", templateName)
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(&buff, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buff.Bytes())
}

// RenderJSON attempts to render an object as JSON. Gives a 404 error on failure
func RenderJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

// RenderTextJSON attempts to render an object as JSON with mime type text/html.
// Gives a 404 error on failure
func RenderTextJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(b)
}

// RenderTextErrorJSON attempts to render an object as JSON with mime type
// text/html and a 404 error.
func RenderTextErrorJSON(w http.ResponseWriter, value interface{}) {
	b, err := json.Marshal(value)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		http.Error(w, "{}", http.StatusNotFound)
		return
	}

	http.Error(w, string(b), http.StatusBadRequest)
}

// URIArgs gets arguments from the current page URI
func URIArgs(r *http.Request) []string {
	args := make([]string, 0, 10)
	numFound := 0

	uri := r.URL.RequestURI()

	var currArg bytes.Buffer
	for i := 0; i < len(uri); i++ {
		if uri[i] == '?' {
			if currArg.Len() > 0 {
				args = append(args, currArg.String())
				currArg.Reset()
				numFound++
			}
			break
		} else if uri[i] == '/' {
			if currArg.Len() > 0 {
				args = append(args, currArg.String())
				numFound++
			}
			currArg.Reset()
		} else {
			currArg.Write([]byte{uri[i]})
		}
	}
	if currArg.Len() > 0 {
		args = append(args, currArg.String())
		numFound++
	}
	return args[:numFound]
}
