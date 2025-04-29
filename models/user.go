package models

import "time"

// User represents user data in the system
type User struct {
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	IsAnonymous   bool   `json:"isAnonymous"`
	Metadata      struct {
		LastSignInTime int `json:"lastSignInTime"`
		CreationTime   int `json:"creationTime"`
	} `json:"metadata"`
	PhoneNumber string `json:"phoneNumber"`
	PhotoURL    string `json:"photoURL"`
	ProviderId  string `json:"providerId"`
	Uid         string `json:"uid"`
}

func (u *User) Save() (User, error) {
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("INSERT INTO users (uid, email, displayName, photoURL, phoneNumber, emailVerified, isAnonymous, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.Uid, u.Email, u.DisplayName, u.PhotoURL, u.PhoneNumber, u.EmailVerified, u.IsAnonymous, u.Metadata)
	if err != nil {
		return User{}, err
	}
	return *u, nil
}
func (u *User) Update(uid string) (User, error) {
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("UPDATE users SET email = ?, displayName = ?, photoURL = ?, phoneNumber = ?, emailVerified = ?, isAnonymous = ?, metadata = ? WHERE uid = ?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.Email, u.DisplayName, u.PhotoURL, u.PhoneNumber, u.EmailVerified, u.IsAnonymous, u.Metadata, uid)
	if err != nil {
		return User{}, err
	}
	return *u, nil
}
func (u *User) Delete(uid string) error {
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("DELETE FROM users WHERE uid = ?")
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
func (u *User) CheckUser(uid string) (User, error) {
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("SELECT * FROM users WHERE uid = ?")
	if err != nil {
		return User{}, err
	}
	defer stmt.Close()
	var user User
	err = stmt.QueryRow(u.Uid).Scan(&user.Uid, &user.Email, &user.DisplayName, &user.PhotoURL, &user.PhoneNumber, &user.EmailVerified, &user.IsAnonymous, &user.Metadata)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
func (u *User) CreateMetadata() error {
	u.Metadata.LastSignInTime = int(time.Now().Unix())
	db := GetDBInstance(GetDBConfig())
	stmt, err := db.Prepare("UPDATE users SET metadata = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.Metadata, u.Uid)
	if err != nil {
		return err
	}
	return nil
}
