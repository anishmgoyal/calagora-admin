package services

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anishmgoyal/calagora-admin/models"
)

var sessions map[string]models.Session
var mutex *sync.Mutex

func initSessions() {
	go sessionEvicterAndDownloader()
	sessions = make(map[string]models.Session)
	mutex = &sync.Mutex{}
}

// sessionEvicterAndDownloader evicts old sessions, and attempts to
// download new emails for active users
func sessionEvicterAndDownloader() {
	for {
		time.Sleep(time.Second * 15)
		var toRemove []string

		emailDownloadQueue := make(map[int]int)

		// Thread-safe session eviction
		mutex.Lock()
		for k, v := range sessions {
			if v.Modified.Unix() < time.Now().AddDate(0, 0, -1).Unix() {
				toRemove = append(toRemove, k)
			} else {
				emailDownloadQueue[v.UserID] = v.UserID
			}
		}
		for _, k := range toRemove {
			delete(sessions, k)
		}
		mutex.Unlock()

		go emailDownloaderForActiveUsers(emailDownloadQueue)
	}
}

func emailDownloaderForActiveUsers(toDownload map[int]int) {
	events := make(chan string)
	go DownloadEmailForUser("support", events)
	for k := range toDownload {
		user, err := models.GetUserByID(Base.DB, k)
		if err != nil || user == nil {
			continue
		}
		<-events
		go DownloadEmailForUser(user.Username, events)
	}
}

func randomSessionID() string {
	bytes := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return ""
	}

	encodedString := base64.StdEncoding.EncodeToString(bytes)
	return encodedString
}

// AddSession attempts to store information about a session
func AddSession(user *models.User, w http.ResponseWriter, r *http.Request) {

	var sessionID string
	for {
		sessionID = randomSessionID()
		mutex.Lock()
		_, ok := sessions[sessionID]
		if !ok {
			break
		}
		mutex.Unlock()
	}
	defer mutex.Unlock()

	sessionSecret := randomSessionID()

	sessions[sessionID] = models.Session{
		ID:       sessionID,
		UserID:   user.ID,
		Secret:   sessionSecret,
		Agent:    r.UserAgent(),
		Modified: time.Now(),
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "SessionID",
		Value:   sessionID,
		Expires: time.Now().AddDate(0, 0, 7),
		Path:    "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "SessionSecret",
		Value:   sessionSecret,
		Expires: time.Now().AddDate(0, 0, 7),
		Path:    "/",
	})
}

// GetSession attempts to pull a session from request data
func GetSession(r *http.Request) *models.Session {
	sessionIDCookie, err := r.Cookie("SessionID")
	if err != nil {
		return nil
	}
	sessionID := sessionIDCookie.Value

	sessionSecretCookie, err := r.Cookie("SessionSecret")
	if err != nil {
		return nil
	}
	sessionSecret := sessionSecretCookie.Value

	mutex.Lock()
	defer mutex.Unlock()

	session, ok := sessions[sessionID]
	if !ok {
		return nil
	}

	if strings.Compare(session.Agent, r.UserAgent()) == 0 &&
		strings.Compare(session.Secret, sessionSecret) == 0 {

		session.Modified = time.Now()
		sessions[sessionID] = session
		return &session
	}

	return nil
}

// DeleteSession facilitates logouts
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionIDCookie, err := r.Cookie("SessionID")
	if err != nil {
		return
	}
	sessionID := sessionIDCookie.Value

	sessionSecretCookie, err := r.Cookie("SessionSecret")
	if err != nil {
		return
	}
	sessionSecret := sessionSecretCookie.Value

	mutex.Lock()
	defer mutex.Unlock()

	session, ok := sessions[sessionID]
	if !ok {
		return
	}

	if strings.Compare(session.Agent, r.UserAgent()) == 0 &&
		strings.Compare(session.Secret, sessionSecret) == 0 {

		http.SetCookie(w, &http.Cookie{
			Name:    "SessionID",
			Expires: time.Now().Add(time.Second * -1),
			Path:    "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:    "SessionSecret",
			Expires: time.Now().Add(time.Second * -1),
			Path:    "/",
		})
		delete(sessions, sessionID)
	}
}
