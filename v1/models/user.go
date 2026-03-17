package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

type User struct {
	ID            int       `json:"id" db:"id"`
	DisplayName   string    `json:"display_name" db:"display_name"`
	Email         string    `json:"email" db:"email"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	IsAnonymous   bool      `json:"is_anonymous" db:"is_anonymous"`
	PhoneNumber   string    `json:"phone_number" db:"phone_number"`
	PhotoURL      string    `json:"photo_url" db:"photo_url"`
	ProviderId    string    `json:"provider_id" db:"provider_id"`
	Uid           string    `json:"uid" db:"firebase_uid"`
	Designation   string    `json:"designation" db:"designation"`
	Branch        string    `json:"branch" db:"branch"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	LoginAt       time.Time `json:"login_at" db:"login_at"`
}

// CreateUser inserts a new user and returns the created record
func CreateUser(user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	if user.Uid == "" {
		return nil, fmt.Errorf("uid (firebase_uid) is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		INSERT INTO users (firebase_uid, email, display_name, photo_url, phone_number, email_verified, is_anonymous, provider_id, login_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	loginAt := time.Now()
	if !user.LoginAt.IsZero() {
		loginAt = user.LoginAt
	}
	_, err := db.Exec(query,
		user.Uid, user.Email, user.DisplayName, user.PhotoURL, user.PhoneNumber,
		user.EmailVerified, user.IsAnonymous, user.ProviderId, loginAt,
	)
	if err != nil {
		return nil, err
	}
	return GetUserByUID(user.Uid)
}

// GetUserByUID retrieves a user by firebase_uid
func GetUserByUID(uid string) (*User, error) {
	if uid == "" {
		return nil, fmt.Errorf("uid is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT firebase_uid, IFNULL(email, ''), IFNULL(display_name, ''), IFNULL(photo_url, ''),
			IFNULL(phone_number, ''), IFNULL(email_verified, 0), IFNULL(is_anonymous, 0),
			IFNULL(provider_id, ''), IFNULL(login_at, NOW())
		FROM users WHERE firebase_uid = ?
	`
	var u User
	u.Uid = uid
	err := db.QueryRow(query, uid).Scan(
		&u.Uid, &u.Email, &u.DisplayName, &u.PhotoURL, &u.PhoneNumber,
		&u.EmailVerified, &u.IsAnonymous, &u.ProviderId, &u.LoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT firebase_uid, IFNULL(email, ''), IFNULL(display_name, ''), IFNULL(photo_url, ''),
			IFNULL(phone_number, ''), IFNULL(email_verified, 0), IFNULL(is_anonymous, 0),
			IFNULL(provider_id, ''), IFNULL(login_at, NOW())
		FROM users WHERE email = ?
	`
	var u User
	err := db.QueryRow(query, email).Scan(
		&u.Uid, &u.Email, &u.DisplayName, &u.PhotoURL, &u.PhoneNumber,
		&u.EmailVerified, &u.IsAnonymous, &u.ProviderId, &u.LoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// GetAllUsers retrieves all users, ordered by login_at desc
func GetAllUsers(limit, offset int) ([]User, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT firebase_uid, IFNULL(email, ''), IFNULL(display_name, ''), IFNULL(photo_url, ''),
			IFNULL(phone_number, ''), IFNULL(email_verified, 0), IFNULL(is_anonymous, 0),
			IFNULL(provider_id, ''), IFNULL(login_at, NOW())
		FROM users ORDER BY login_at DESC
	`
	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	var rows *sql.Rows
	var err error
	if len(args) > 0 {
		rows, err = db.Query(query, args...)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(
			&u.Uid, &u.Email, &u.DisplayName, &u.PhotoURL, &u.PhoneNumber,
			&u.EmailVerified, &u.IsAnonymous, &u.ProviderId, &u.LoginAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates an existing user by firebase_uid
func UpdateUser(user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.Uid == "" {
		return fmt.Errorf("uid is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		UPDATE users SET
			email = ?, display_name = ?, photo_url = ?, phone_number = ?,
			email_verified = ?, is_anonymous = ?, provider_id = ?, login_at = ?
		WHERE firebase_uid = ?
	`
	result, err := db.Exec(query,
		user.Email, user.DisplayName, user.PhotoURL, user.PhoneNumber,
		user.EmailVerified, user.IsAnonymous, user.ProviderId, user.LoginAt,
		user.Uid,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// DeleteUser deletes a user by firebase_uid
func DeleteUser(uid string) error {
	if uid == "" {
		return fmt.Errorf("uid is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `DELETE FROM users WHERE firebase_uid = ?`
	result, err := db.Exec(query, uid)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user not found")
	}
	return nil
}
// GetUserByID retrieves a user by numeric id
func GetUserByID(id int) (*User, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT id, firebase_uid, IFNULL(email, ''), IFNULL(display_name, ''), IFNULL(photo_url, ''),
			IFNULL(phone_number, ''), IFNULL(email_verified, 0), IFNULL(is_anonymous, 0),
			IFNULL(provider_id, ''), IFNULL(login_at, NOW())
		FROM users WHERE id = ?
	`
	var u User
	err := db.QueryRow(query, id).Scan(
		&u.ID, &u.Uid, &u.Email, &u.DisplayName, &u.PhotoURL, &u.PhoneNumber,
		&u.EmailVerified, &u.IsAnonymous, &u.ProviderId, &u.LoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}