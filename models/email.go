package models

import (
	"database/sql"
	"strings"
	"time"
)

const (
	// RecipientTypeTo is any original recipient of an email
	RecipientTypeTo = "to"
	// RecipientTypeCC is any secondary recipient of an email
	RecipientTypeCC = "cc"
	// RecipientTypeBCC is any hidden recipient of an email
	RecipientTypeBCC = "bcc"
	// EmailPageSize represents how many emails are loaded per page
	EmailPageSize = 50
)

type emailPair struct {
	DisplayName  string
	EmailAddress string
}

// Email encapsulates any information needed to render an
// email message
type Email struct {
	ID             int          `json:"id"`
	To             []string     `json:"to"`
	CC             []string     `json:"cc"`
	BCC            []string     `json:"bcc"`
	From           string       `json:"from"`
	FromName       string       `json:"from_name"`
	Subject        string       `json:"subject"`
	PlainText      string       `json:"plain_text"`
	FormattedText  string       `json:"formatted_text"`
	HasAttachments bool         `json:"has_attachments"`
	Read           bool         `json:"is_read"`
	IsSpam         bool         `json:"is_spam"`
	IsVirus        bool         `json:"is_virus"`
	Attachments    []Attachment `json:"attachments"`
	Received       time.Time    `json:"received"`
}

// Create attempts to add an email to the database
func (e *Email) Create(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	fromPair := parseEmailAddress(e.From)
	rows, err := tx.Query("INSERT INTO emails (from_display, from_addr, "+
		"subject, plain_text, formatted_text, is_spam, is_virus, received) VALUES "+
		"($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", fromPair.DisplayName,
		fromPair.EmailAddress, e.Subject, e.PlainText, e.FormattedText, e.IsSpam,
		e.IsVirus, e.Received)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows.Next() {
		err = rows.Scan(&e.ID)
		rows.Close()
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		rows.Close()
	}

	for i := 0; i < len(e.Attachments); i++ {
		a := &e.Attachments[i]
		a.EmailID = e.ID
		err = a.Create(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = e.addRecipients(tx, e.To, RecipientTypeTo); err != nil {
		tx.Rollback()
		return err
	}
	if err = e.addRecipients(tx, e.CC, RecipientTypeCC); err != nil {
		tx.Rollback()
		return err
	}
	if err = e.addRecipients(tx, e.BCC, RecipientTypeBCC); err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func (e *Email) addRecipients(tx *sql.Tx, emailAddresses []string,
	recipientType string) error {

	for _, emailAddress := range emailAddresses {
		if len(emailAddress) > 3 {
			if err := e.addRecipient(tx, emailAddress, RecipientTypeTo); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Email) addRecipient(tx *sql.Tx, emailAddress string,
	recipientType string) error {

	_, err := tx.Exec("INSERT INTO recipients (email_id, email_address, "+
		"recipient_type) VALUES ($1, $2, $3)", e.ID, emailAddress,
		recipientType)
	return err
}

// Delete attempts to remove an email from the database
func (e *Email) Delete(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = e.GetAttachmentsForEmail(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, attachment := range e.Attachments {
		err = attachment.Delete(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM recipients WHERE email_id = $1", e.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM emails WHERE email_id = $1", e.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

// GetEmailByID attempts to load an email into memory and return it
func GetEmailByID(db *sql.DB, id int) (*Email, error) {
	// Build the base email struct
	rows, err := db.Query("SELECT from_display, from_addr, subject, plain_text, "+
		"formatted_text, is_spam, is_virus, received FROM emails WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	email := Email{ID: id}
	err = rows.Scan(&email.FromName, &email.From, &email.Subject,
		&email.PlainText, &email.FormattedText, &email.IsSpam, &email.IsVirus,
		&email.Received)
	if err != nil {
		return nil, err
	}

	// Load in recipients
	email.To = make([]string, 0, 10)
	email.CC = make([]string, 0, 10)
	email.BCC = make([]string, 0, 10)

	rows2, err := db.Query("SELECT email_address, recipient_type FROM "+
		"recipients WHERE email_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var emailAddress, recipientType string
		err = rows2.Scan(&emailAddress, &recipientType)
		if err != nil {
			return nil, err
		}
		switch recipientType {
		case "to":
			email.To = append(email.To, emailAddress)
		case "cc":
			email.CC = append(email.CC, emailAddress)
		case "bcc":
			email.BCC = append(email.BCC, emailAddress)
		}
	}

	// Load in attachment data
	err = email.GetAttachmentsForEmail(db)
	if err != nil {
		return nil, err
	}

	return &email, nil
}

// IsRecipient determines if a user received an email
func (u *User) IsRecipient(db *sql.DB, emailID int) (bool, error) {
	row := db.QueryRow("SELECT count(1) FROM recipients WHERE email_id = "+
		"$1 AND email_address = $2", emailID, u.EmailAddress)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkRead attempts to mark an email read
func (e *Email) MarkRead(db *sql.DB) error {
	_, err := db.Exec("UPDATE emails set is_read = true WHERE id = $1", e.ID)
	return err
}

// GetEmailSendCount attempts to find the number of emails a user has sent
func (u *User) GetEmailSendCount(db *sql.DB) int {
	rows, err := db.Query("SELECT count(id) FROM emails WHERE from_addr "+
		" = $1", u.EmailAddress)
	if err != nil {
		return 0
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if rows.Scan(&count) == nil {
			return count
		}
	}
	return 0
}

// GetEmailCount gets the number of emails a user has received
func (u *User) GetEmailCount(db *sql.DB) int {
	rows, err := db.Query("SELECT count(email_id) FROM recipients WHERE "+
		"email_address = $1", u.EmailAddress)
	if err != nil {
		return 0
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if rows.Scan(&count) == nil {
			return count
		}
	}
	return 0
}

// LoadEmailPage gets a page of emails for a user
func (u *User) LoadEmailPage(db *sql.DB, page int) ([]Email, error) {
	emails, err := LoadEmailsForAddress(db, u.EmailAddress, page)
	return emails, err
}

// LoadEmailsForAddress attempts to get stubs for emails sent to a user
func LoadEmailsForAddress(db *sql.DB, emailAddress string, page int) ([]Email,
	error) {

	rows, err := db.Query("SELECT id, from_display, from_addr, subject, "+
		"plain_text, formatted_text, is_read, is_spam, is_virus, received, "+
		"(exists(SELECT * FROM attachments WHERE email_id = e.id)) has_attachments "+
		"FROM emails e WHERE exists(SELECT * FROM recipients WHERE "+
		"email_id = e.id AND email_address = $1) ORDER BY received DESC LIMIT $2 "+
		"OFFSET $3", emailAddress, EmailPageSize, EmailPageSize*page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := make([]Email, 0, 50)
	for rows.Next() {
		var email Email
		err = rows.Scan(&email.ID, &email.FromName, &email.From, &email.Subject,
			&email.PlainText, &email.FormattedText, &email.Read, &email.IsSpam,
			&email.IsVirus, &email.Received, &email.HasAttachments)
		if err == nil {
			emails = append(emails, email)
		}
	}

	return emails, nil
}

func parseEmailAddress(emailString string) emailPair {
	idxStart := strings.Index(emailString, "<")
	idxEnd := strings.Index(emailString, ">")
	pair := emailPair{}
	if idxEnd > idxStart {
		pair.DisplayName = strings.TrimSpace(emailString[0:idxStart])
		pair.EmailAddress = strings.TrimSpace(emailString[idxStart+1 : idxEnd])
	} else {
		pair.DisplayName = emailString
		pair.EmailAddress = emailString
	}
	return pair
}
