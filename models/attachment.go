package models

import "database/sql"

// Attachment encapsulates any data needed to keep track of
// an email attachment
type Attachment struct {
	ID          int    `json:"id"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	RawData     []byte `json:"-"`
	EmailID     int    `json:"email_id"`
}

// Create attempts to save information about an attachment to the database
func (a *Attachment) Create(db dbInterface) error {
	rows, err := db.Query("INSERT INTO attachments (content_type, file_name, "+
		"file_path, email_id) VALUES ($1, $2, $3, $4) RETURNING id", a.ContentType,
		a.FileName, a.FilePath, a.EmailID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&a.ID)
	}
	return nil
}

// Save attempts to save information about an attachment to the database
func (a *Attachment) Save(db dbInterface) error {
	_, err := db.Exec("UPDATE attachments SET content_type = $1, "+
		"file_name = $2, file_path = $3, email_id = $4 WHERE id = $5",
		a.ContentType, a.FileName, a.FilePath, a.EmailID, a.ID)
	return err
}

// Delete attempts to remove an attachment from the database
func (a *Attachment) Delete(db dbInterface) error {
	_, err := db.Exec("DELETE FROM attachments WHERE id = $1", a.ID)
	return err
}

// GetAttachmentByID tries to find an attachment by its ID
func GetAttachmentByID(db *sql.DB, id int) (*Attachment, error) {
	rows, err := db.Query("SELECT id, content_type, file_name, file_path, "+
		"email_id FROM attachments WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var attachment Attachment
		err = rows.Scan(&attachment.ID, &attachment.ContentType,
			&attachment.FileName, &attachment.FilePath, &attachment.EmailID)
		if err == nil {
			return &attachment, nil
		}
	}

	return nil, nil
}

// GetAttachmentsForEmail attempts to build data about attachments for
// a given email, and insert them into the email object
func (e *Email) GetAttachmentsForEmail(db dbInterface) error {
	rows, err := db.Query("SELECT id, content_type, file_name, file_path "+
		"FROM attachments WHERE email_id = $1", e.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	e.Attachments = make([]Attachment, 0, 10)
	for rows.Next() {
		attachment := Attachment{EmailID: e.ID}
		err = rows.Scan(&attachment.ID, &attachment.ContentType,
			&attachment.FileName, &attachment.FilePath)
		if err == nil {
			e.Attachments = append(e.Attachments, attachment)
		}
	}
	return nil
}
