package bootstrap

import (
	"net/http"

	"github.com/anishmgoyal/calagora-admin/controllers"
	"github.com/anishmgoyal/calagora-admin/resources"
)

// CreateRoutes maps URI's to the corresponding controller method
func CreateRoutes() {
	resources.MapCSSHandler()
	resources.MapImageHandler()
	resources.MapJSHandler()

	http.Handle(route("/", controllers.Home))
	http.Handle(route("/logout/", controllers.Logout))

	http.Handle(route("/attachment/", controllers.Attachment))

	http.Handle(route("/email/view/", controllers.EmailView))
	http.Handle(route("/email/", controllers.Email))
}

// Quick wrapper for StripPrefix which prevents typos
func route(path string, callback http.HandlerFunc) (string, http.Handler) {
	return path, http.StripPrefix(path, callback)
}
