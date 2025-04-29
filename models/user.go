package models

import (
	"fmt"
	"time"
)

// User represents user data in the system
type User struct {
	DisplayName   string    `json:"displayName"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	IsAnonymous   bool      `json:"isAnonymous"`
	PhoneNumber   string    `json:"phoneNumber"`
	PhotoURL      string    `json:"photoURL"`
	ProviderId    string    `json:"providerId"`
	Uid           string    `json:"uid"`
	CreatedAt     time.Time `json:"created_at"`
	LoginAt       time.Time `json:"login_at"`
}

func (u *User) Save() (User, error) {
	// save user
	fmt.Println("---USER SAVE---")
	db := GetDBInstance(GetDBConfig())
	loginAt := time.Now()
	stmt, err := db.Prepare("INSERT INTO users (firebase_uid, email, display_name, photo_url, phone_number, email_verified, is_anonymous, provider_id, login_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.Uid, u.Email, u.DisplayName, u.PhotoURL, u.PhoneNumber, u.EmailVerified, u.IsAnonymous, u.ProviderId, loginAt)
	fmt.Println("@@@save err:", err)
	if err != nil {
		fmt.Println("Save error:", err)
		return User{}, err
	}
	return *u, nil
}
func (u *User) Update(uid string) (User, error) {
	fmt.Println("---USER UPDATE---")
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("UPDATE users SET email = ?, display_name = ?, photo_url = ?, phone_number = ?, email_verified = ?, is_anonymous = ?, provider_id, login_at = ? WHERE firebase_uid = ?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.Email, u.DisplayName, u.PhotoURL, u.PhoneNumber, u.EmailVerified, u.IsAnonymous, u.ProviderId, u.LoginAt)
	if err != nil {
		return User{}, err
	}
	return *u, nil
}
func (u *User) Delete(uid string) error {
	fmt.Println("---USER DELETE---")
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("DELETE FROM users WHERE firebase_uid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uid)
	if err != nil {
		return err
	}
	return nil
}
func (u *User) IsUserSaved(uid string) bool {
	fmt.Println("---USER CHECK---")
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("SELECT firebase_uid FROM users WHERE firebase_uid = ?")
	if err != nil {
		fmt.Println("IsUserSaved error:", err)
		return false
	}
	defer stmt.Close()
	var user User
	err = stmt.QueryRow(u.Uid).Scan(
		&user.Uid,
	)
	if err != nil {
		fmt.Println("IsUserSaved error:", err)
		return false
	}
	fmt.Println("IsUserSaved result:", user)
	return true
}
