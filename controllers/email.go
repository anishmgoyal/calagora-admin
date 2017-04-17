package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora-admin/models"
)

type emailViewData struct {
	EmailAccountName string
	CurrentAddress   string
	ActiveSelector   string
	Page             int
	SwitchToName     string
	SwitchToHref     string
	Emails           []models.Email
}

const (
	pageInbox = iota
	pageSent
	pageJunk
	pageTrash
	numPages
)

const (
	supportEmail = "support@calagora.com"
	emailSuffix  = "@calagora.com"
)

var selectors []string

func emailInit() {
	selectors = make([]string, numPages)
	selectors[pageInbox] = "#inbox"
	selectors[pageSent] = "#sent"
	selectors[pageJunk] = "#junk"
	selectors[pageTrash] = "#trash"
}

func getPage(selector string) int {
	switch selector {
	case "inbox":
		return pageInbox
	case "sent":
		return pageSent
	case "junk":
		return pageJunk
	case "trash":
		return pageTrash
	default:
		return pageInbox
	}
}

// Email handles the route '/email/'
func Email(w http.ResponseWriter, r *http.Request) {
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

	if len(args) < 2 {
		args = append(args, "inbox")
	}

	data := &emailViewData{
		EmailAccountName: args[0],
		Page:             getPage(args[1]),
	}
	data.ActiveSelector = selectors[data.Page]

	user, err := models.GetUserByID(Base.Db, viewData.Session.UserID)
	if err != nil {
		user = nil
	}

	emailUser := models.User{}

	if strings.Compare(args[0], "mine") == 0 {
		// Render my email account
		if user != nil {
			data.CurrentAddress = user.Username + emailSuffix
			emailUser.EmailAddress = user.EmailAddress
		} else {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		data.SwitchToName = supportEmail
		data.SwitchToHref = "/email/support/"
	} else if strings.Compare(args[0], "support") == 0 {
		// Render support email account
		data.CurrentAddress = supportEmail
		emailUser.EmailAddress = supportEmail
		if user != nil {
			data.SwitchToName = user.Username + emailSuffix
		} else {
			data.SwitchToName = "Your Email"
		}
		data.SwitchToHref = "/email/mine/"
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	emails, err := emailUser.LoadEmailPage(Base.Db, 0)
	if err == nil {
		data.Emails = emails
	} else {
		data.Emails = []models.Email{}
	}

	viewData.Data = data

	RenderView(w, "email#index", viewData)
}

type emailViewViewData struct {
	Email            models.Email
	EmailAccountName string
	CurrentAddress   string
	SwitchToName     string
	SwitchToHref     string
	Emails           []models.Email
}

// EmailView renders the route '/email/view/#id'
func EmailView(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	if viewData.Session == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.GetUserByID(Base.Db, viewData.Session.UserID)
	if err != nil || user == nil {
		http.Error(w, "Error", http.StatusNotFound)
		return
	}

	args := URIArgs(r)
	if len(args) < 2 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	data := &emailViewViewData{
		EmailAccountName: args[0],
	}

	idStr := args[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	email, err := models.GetEmailByID(Base.Db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if email == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if strings.Compare(args[0], "mine") == 0 {
		data.CurrentAddress = user.Username + emailSuffix
		data.SwitchToHref = "/email/support/"
		data.SwitchToName = supportEmail
	} else {
		data.CurrentAddress = supportEmail
		data.SwitchToHref = "/email/mine/"
		data.SwitchToName = user.Username + emailSuffix
	}

	found := false
	recipients := make([]string, len(email.To)+len(email.CC)+len(email.BCC))
	offset := 0
	for _, slice := range [][]string{email.To, email.CC, email.BCC} {
		copy(recipients[offset:], slice)
		offset += len(slice)
	}

	for _, recipient := range recipients {
		if strings.Compare(recipient, user.EmailAddress) == 0 ||
			strings.Compare(recipient, "support@calagora.com") == 0 {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data.Email = *email
	viewData.Data = data

	email.MarkRead(Base.Db)

	RenderView(w, "email#view", viewData)
}
