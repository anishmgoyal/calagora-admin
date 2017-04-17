package models

import "time"

// Session encapsulates session data. Stored in-memory
type Session struct {
	ID       string
	UserID   int
	Secret   string
	Agent    string
	Modified time.Time
}
