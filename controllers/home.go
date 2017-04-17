package controllers

import (
	"net/http"

	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/anishmgoyal/calagora-admin/services"
)

type homeData struct {
	Error bool
}

// Logout handles '/logout/'
func Logout(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	services.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Home is the index page for the admin panel
func Home(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if r.Method == http.MethodPost {
		postHome(w, r)
		return
	}
	viewData.Data = &homeData{
		Error: false,
	}
	RenderView(w, "home#index", viewData)
}

func postHome(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := models.GetUserByUsername(Base.Db, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Authenticate(password) {
		services.AddSession(user, w, r)
		go services.DownloadEmailForUser(user.Username, make(chan string))

		viewData.Data = &homeData{
			Error: false,
		}
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		viewData.Data = &homeData{
			Error: true,
		}
		RenderView(w, "home#index", viewData)
	}
}
