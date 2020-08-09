package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
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

func (u UserService) Authenticate(username, password string) (bool, interface{}) {
	user, operr := u.FindByEmail(username)
	if operr != nil {
		log.Printf("error while authenticating: %v", operr)
		return false, nil
	}

	if user.Password == password {
		return true, user
	}

	return false, nil
}

func (u UserService) Register(user User) (string, *DBOpError) {
	db := u.DB
	tx := db.MustBegin()
	sql := `
		INSERT INTO 
			users (user_id, platform_name, email, passhash, firstname, lastname, created_date) 
		VALUES 
			(:user_id, :platform_name, :email, :passhash, :firstname, :lastname, :created_date)
	`

	user.ID = uuid.New().String()
	user.Created = CurrentTimeInSeconds()
	user.PlatformName = "web"
	_, err := tx.NamedExec(sql, &user)
	if err != nil {
		tx.Rollback()
		return "", &DBOpError{sql, err}
	}

	err = tx.Commit()
	if err != nil {
		return "", &DBOpError{sql, err}
	}

	return user.ID, nil
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

func (u UserService) FindByEmail(email string) (User, *DBOpError) {
	db := u.DB
	user := User{}
	sql := "SELECT * FROM users WHERE email=$1"
	err := db.Get(&user, sql, email)
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

func (u UserService) Update(user User) *DBOpError {
	db := u.DB
	tx := db.MustBegin()
	sql := `
		UPDATE 
			users 
		SET 
			firstname      = :firstname,  
		    lastname       = :lastname,
		    mfa_enabled     = :mfa_enabled 
		WHERE 
			user_id = :user_id 
	`

	_, err := tx.NamedExec(sql, &user)
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

func (u UserService) UpdateMFATemp(user User) *DBOpError {
	db := u.DB
	tx := db.MustBegin()
	sql := `
		UPDATE 
			users 
		SET 
		    mfa_secret_tmp        = :mfa_secret_tmp, 
		    mfa_secret_tmp_exp    = :mfa_secret_tmp_exp
		WHERE 
			user_id = :user_id 
	`

	_, err := tx.NamedExec(sql, &user)
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

func (u UserService) UpdateMFA(user User) *DBOpError {
	db := u.DB
	tx := db.MustBegin()
	sql := `
		UPDATE 
			users 
		SET 
		    mfa_secret_current    = :mfa_secret_current, 
		    mfa_secret_tmp        = :mfa_secret_tmp, 
		    mfa_secret_tmp_exp    = :mfa_secret_tmp_exp,
		    mfa_enabled           = :mfa_enabled
		WHERE 
			user_id = :user_id 
	`

	log.Printf("sql = \n%v", sql)

	_, err := tx.NamedExec(sql, &user)
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
