package controllers

import (
	"net/http"
	"strconv"

	"github.com/anishmgoyal/calagora-admin/constants"
	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/anishmgoyal/calagora-admin/utils"
)

// Attachment handles the route '/attachment/'
func Attachment(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	args := URIArgs(r)
	if len(args) < 1 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	attachment, err := models.GetAttachmentByID(Base.Db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if attachment == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	user, err := models.GetUserByID(Base.Db, viewData.Session.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	supportUser := models.User{EmailAddress: constants.SupportEmail}

	ok, err := user.IsRecipient(Base.Db, attachment.EmailID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		ok, err = supportUser.IsRecipient(Base.Db, attachment.EmailID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	body, err := utils.LoadAttachment(*attachment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", attachment.ContentType)
	w.Write(body)
}
