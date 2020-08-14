package dbutil

import (
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/jmoiron/sqlx"
	"time"
)

type UserService struct {
	*sqlx.DB
}

func CurrentTimeInSeconds() int64 {
	return time.Now().Unix()
}

func (u UserService) Upsert(sql string, obj interface{}) *DBOpError {
	db := u.DB
	tx := db.MustBegin()

	_, err := tx.NamedExec(sql, obj)
	if err != nil {
		tx.Rollback()
		return &DBOpError{sql, err}
	}

	err = tx.Commit()
	if err != nil {
		return &DBOpError{sql, err}
	}

	return nil
}

func (u UserService) Register(user User) *DBOpError {
	sql := `
		INSERT INTO users
			  (user_id,  platform_name,  email,   passhash,  firstname, lastname,  created_date) 
		VALUES 
			  (:user_id, :platform_name, :email, :passhash, :firstname, :lastname, :created_date)
	`
	return u.Upsert(sql, &user)
}

func (u UserService) Get(id string) (User, *DBOpError) {
	db := u.DB
	user := User{}
	sql := "SELECT * FROM users WHERE user_id=$1"
	err := db.Get(&user, sql, id)
	if err != nil {
		return user, &DBOpError{sql, err}
	}
	return user, nil
}

func (u UserService) Delete(id string) *DBOpError {
	db := u.DB
	sql := "DELETE FROM users WHERE user_id=$1"
	_, err := db.Exec(sql, id)
	if err != nil {
		return &DBOpError{sql, err}
	}
	return nil
}

func (u UserService) FindByEmail(email string) (*User, *DBOpError) {
	db := u.DB
	user := &User{}
	sql := "SELECT * FROM users WHERE email=$1"
	err := db.Get(user, sql, email)
	if err != nil {
		return user, &DBOpError{sql, err}
	}
	return user, nil
}

func (u UserService) FindByEventID(eventID string) (*User, *DBOpError) {
	db := u.DB
	user := &User{}
	sql := `
		SELECT 
				* 
		FROM 
				users
		WHERE 
				user_id  = UUID(TRIM(( SELECT user_id FROM auth_event WHERE event_id=$1 )))
`
	err := db.Get(user, sql, eventID)
	if err != nil {
		return user, &DBOpError{sql, err}
	}
	return user, nil
}

func (u UserService) List() ([]User, *DBOpError) {
	db := u.DB
	users := []User{}
	sql := "SELECT * FROM users"
	err := db.Select(&users, sql)
	if err != nil {
		return users, &DBOpError{sql, err}
	}
	return users, nil
}

func (u UserService) UpdateProfile(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
			firstname      = :firstname,  
		    lastname       = :lastname,
			email          = :email,
			passhash       = :passhash,
			totp_enabled   = :totp_enabled
		WHERE 
			user_id = :user_id 
	`
	return u.Upsert(sql, &user)
}

func (u UserService) UpdateTOTPTmp(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    totp_secret_tmp        = :totp_secret_tmp, 
		    totp_secret_tmp_exp    = :totp_secret_tmp_exp
		WHERE 
			user_id = :user_id 
	`
	return u.Upsert(sql, &user)
}

func (u UserService) UpdateTOTP(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    totp_enabled           = :totp_enabled,
		    totp_secret_current    = :totp_secret_current, 
		    totp_secret_tmp        = :totp_secret_tmp, 
		    totp_secret_tmp_exp    = :totp_secret_tmp_exp
		WHERE 
			user_id = :user_id 
	`
	return u.Upsert(sql, &user)
}

func (u UserService) RecordAuthEvent(auth AuthEvent) *DBOpError {
	sql := `
		INSERT INTO auth_event
			  (event_id,  user_id,  event,   event_timestamp,  ip_v4, ip_v6,  agent) 
		VALUES 
			  (:event_id, :user_id, :event, :event_timestamp, :ip_v4, :ip_v6, :agent)
	`
	return u.Upsert(sql, &auth)
}
