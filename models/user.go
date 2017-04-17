package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// User encapsulates user data
type User struct {
	ID                int
	Username          string
	EmailAddress      string
	Password          string
	EncryptedPassword string
	Salt              string
}

func (u *User) encryptPassword(salt []byte) error {
	if salt == nil {
		salt = make([]byte, 32)
		_, err := io.ReadFull(rand.Reader, salt)
		if err != nil {
			return errors.New("Failed to generate salt")
		}
	}

	encryptedPassword, err := scrypt.Key([]byte(u.Password), salt, 16384, 8,
		1, 32)
	if err != nil {
		return errors.New("Failed to encrypt")
	}

	encodedPassword := base64.StdEncoding.EncodeToString(encryptedPassword)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	u.Salt = encodedSalt
	u.EncryptedPassword = encodedPassword

	return nil
}

func comparePassword(sourcePassword string,
	encodedPassword string,
	encodedSalt string) bool {

	salt, err := base64.StdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false
	}

	encryptedPassword, err := scrypt.Key([]byte(sourcePassword), salt, 16384, 8,
		1, 32)
	if err != nil {
		return false
	}

	encodedAttempt := base64.StdEncoding.EncodeToString(encryptedPassword)

	return strings.Compare(encodedAttempt, encodedPassword) == 0
}

// Create adds a user to the db
func (u *User) Create(db *sql.DB) error {
	if err := u.encryptPassword(nil); err != nil {
		return err
	}

	rows, err := db.Query("INSERT INTO admusers (username, email_address, "+
		"password, salt) VALUES ($1, $2, $3, $4) RETURNING id", u.Username,
		u.EmailAddress, u.EncryptedPassword, u.Salt)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&u.ID)
	}
	return nil
}

// Authenticate checks if a given password is correct
func (u *User) Authenticate(password string) bool {
	forceFail := false
	if u == nil {
		forceFail = true
		u = &User{}
	}
	if comparePassword(password, u.EncryptedPassword, u.Salt) {
		if !forceFail {
			return true
		}
	}
	return false
}

// GetUserByID tries to get data about a user from the database by id
func GetUserByID(db *sql.DB, id int) (*User, error) {
	user, err := userGetter(db, "id = $1", []interface{}{id})
	return user, err
}

// GetUserByUsername tries to get data about a user from the database by uname
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user, err := userGetter(db, "username = $1", []interface{}{username})
	return user, err
}

func userGetter(db *sql.DB, where string, args []interface{}) (*User, error) {
	rows, err := db.Query("SELECT id, username, email_address, password, salt "+
		"FROM admusers WHERE "+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Username, &user.EmailAddress,
			&user.EncryptedPassword, &user.Salt)
		if err == nil {
			return &user, nil
		}
	}
	return nil, nil
}
