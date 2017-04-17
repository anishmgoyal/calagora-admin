package bootstrap

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anishmgoyal/calagora-admin/constants"
	"github.com/anishmgoyal/calagora-admin/controllers"
	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/anishmgoyal/calagora-admin/services"
)

// GlobalStart begins initialization for the application,
// and notifies main() if an error occurrs.
func GlobalStart() bool {

	fmt.Println("[STARTUP] Loading Environment Settings")
	constants.LoadEnvironmentSettings()

	// Seed for random generators
	rand.Seed(time.Now().UTC().UnixNano())

	fmt.Println("[STARTUP] Loading Templates")
	templates := GetTemplates()

	fmt.Println("[STARTUP] Connecting to DB")
	db := GetDatabaseConnection()

	fmt.Println("[STARTUP] Initializing Services")
	controllers.BaseInitialization(templates, db)
	services.BaseInitialization(db)

	fmt.Println("[STARTUP] Creating Routes")
	CreateRoutes()

	user, err := models.GetUserByUsername(db, "admin")
	if user == nil && err == nil {
		user := models.User{
			Username:     "admin",
			Password:     "ADM@cal2016!!",
			EmailAddress: "admin@calagora.com",
		}
		user.Create(db)
	}

	var sslRedirect = func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if index := strings.Index(host, ":"); index > -1 {
			host = host[0:index]
		}
		redirectURL := "https://" + host
		if constants.SSLPortNum != 443 {
			redirectURL += ":" + strconv.Itoa(constants.SSLPortNum)
		}
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	}

	if constants.SSLEnable {
		fmt.Println("[STARTUP] Starting server on port " +
			strconv.Itoa(constants.SSLPortNum))
		fmt.Println("[STARTUP] Using redirect from port " +
			strconv.Itoa(constants.PortNum))

		go http.ListenAndServe(":"+strconv.Itoa(constants.PortNum),
			http.HandlerFunc(sslRedirect))
		http.ListenAndServeTLS(":"+strconv.Itoa(constants.SSLPortNum),
			constants.SSLCertificate, constants.SSLKeyFile, nil)
	} else {
		fmt.Println("[STARTUP] Starting server on port " +
			strconv.Itoa(constants.PortNum))

		http.ListenAndServe(":"+strconv.Itoa(constants.PortNum), nil)
	}

	// Shouldn't return, I don't believe..
	fmt.Println("[STARTUP] Startup failed.")
	return false
}
