package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

type DBOpError struct {
	Query string
	Err   error
}

type ExecutionError struct {
	ErrCode int    `json: "code"`
	Message string `json: "message"`
}

func NewExecutionError(errCode int, message, txid string) gin.H {
	return gin.H{
		"message": message,
		"code":    errCode,
		"txid":    txid,
	}
}

func (e *DBOpError) Unwrap() error {
	return e.Err
}

type UserService struct {
	*sqlx.DB
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

func CurrentTimeInSeconds() int64 {
	return time.Now().Unix()
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
			mfa_enabled    = :mfa_enabled
		WHERE 
			user_id = :user_id 
	`
	return u.Upsert(sql, &user)
}

func (u UserService) UpdateMFATemp(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    mfa_secret_tmp        = :mfa_secret_tmp, 
		    mfa_secret_tmp_exp    = :mfa_secret_tmp_exp
		WHERE 
			user_id = :user_id 
	`
	return u.Upsert(sql, &user)
}

func (u UserService) UpdateMFA(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    mfa_enabled           = :mfa_enabled,
		    mfa_secret_current    = :mfa_secret_current, 
		    mfa_secret_tmp        = :mfa_secret_tmp, 
		    mfa_secret_tmp_exp    = :mfa_secret_tmp_exp
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

func NewAuthEvent(userID, event, ipv4, ipv6, agent string) AuthEvent {
	return AuthEvent{
		ID:        uuid.New().String(),
		UserIDReq: userID,
		Event:     event,
		IPV4:      ipv4,
		IPV6:      ipv6,
		Agent:     agent,
		Timestamp: CurrentTimeInSeconds(),
	}
}
