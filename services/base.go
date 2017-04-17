package services

import "database/sql"

// Base contains information needed by all or most or many services
var Base struct {
	// DB is the handle to the local database
	DB *sql.DB
}

// BaseInitialization sets up services
func BaseInitialization(db *sql.DB) {
	Base.DB = db
	initSessions()
}
