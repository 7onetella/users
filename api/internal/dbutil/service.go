package dbutil

import (
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/internal/model/oauth2"
	"github.com/jmoiron/sqlx"
	"time"
)

type UserService struct {
	*sqlx.DB
}

func CurrentTimeInSeconds() int64 {
	return time.Now().Unix()
}

func (u UserService) NamedExec(sql string, obj interface{}) *DBOpError {
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

func (u UserService) MustExec(sql string, args ...interface{}) *DBOpError {
	db := u.DB
	tx := db.MustBegin()

	tx.MustExec(sql, args...)

	err := tx.Commit()
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
	return u.NamedExec(sql, &user)
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

func (u UserService) FindUserByAuthEventID(eventID string) (User, *DBOpError) {
	db := u.DB
	user := User{}
	sql := `
		SELECT 
				* 
		FROM 
				users
		WHERE 
				user_id  = UUID(TRIM(( SELECT user_id FROM auth_event WHERE event_id=$1 )))
`
	err := db.Get(&user, sql, eventID)
	if err != nil {
		return user, &DBOpError{sql, err}
	}
	return user, nil
}

func (u UserService) GetAuthEvent(id string) (AuthEvent, *DBOpError) {
	db := u.DB
	auth := AuthEvent{}
	sql := "SELECT * FROM auth_event WHERE event_id=$1"
	err := db.Get(&auth, sql, id)
	if err != nil {
		return auth, &DBOpError{sql, err}
	}
	return auth, nil
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
			firstname        = :firstname,  
		    lastname         = :lastname,
			email            = :email,
			passhash         = :passhash,
			totp_enabled     = :totp_enabled,
            webauthn_enabled = :webauthn_enabled
		WHERE 
			user_id = :user_id 
	`
	return u.NamedExec(sql, &user)
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
	return u.NamedExec(sql, &user)
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
	return u.NamedExec(sql, &user)
}

func (u UserService) RecordAuthEvent(auth AuthEvent) *DBOpError {
	sql := `
		INSERT INTO auth_event
			  (event_id,  user_id,  event,   event_timestamp,  ip_v4, ip_v6,  agent) 
		VALUES 
			  (:event_id, :user_id, :event, :event_timestamp, :ip_v4, :ip_v6, :agent)
	`
	return u.NamedExec(sql, &auth)
}

func (u UserService) SaveWebauthnSession(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    webauthn_session    = :webauthn_session
		WHERE 
			user_id = :user_id 
	`
	return u.NamedExec(sql, &user)
}

func (u UserService) SaveUserCredential(credential UserCredential) *DBOpError {
	sql := `
		INSERT INTO user_credential
			  (credential_id,   user_id,  credential) 
		VALUES 
			  (:credential_id, :user_id, :credential)
	`
	return u.NamedExec(sql, &credential)
}

func (u UserService) UpdateWebAuthn(user User) *DBOpError {
	sql := `
		UPDATE 
			users 
		SET 
		    webauthn_enabled       = :webauthn_enabled,
		    webauthn_session       = :webauthn_session 
		WHERE 
			user_id = :user_id 
	`
	return u.NamedExec(sql, &user)
}

func (u UserService) FindUserCredentialByUserID(userID string) ([]UserCredential, *DBOpError) {
	db := u.DB
	sql := `
		SELECT 
				* 
		FROM 
				user_credential
		WHERE 
				user_id  = $1
`
	credentials := []UserCredential{}

	err := db.Select(&credentials, sql, userID)
	if err != nil {
		return nil, &DBOpError{sql, err}
	}

	return credentials, nil
}

func (u UserService) DoesClientExist(clientID string) (bool, *DBOpError) {
	client, dberr := u.GetClient(clientID)
	if dberr != nil {
		return false, dberr
	}

	return len(client.ID) > 0 && len(client.Name) > 0, nil
}

func (u UserService) GetClient(clientID string) (oauth2.Client, *DBOpError) {
	db := u.DB
	sql := `SELECT * FROM clients WHERE client_id  = $1`
	client := oauth2.Client{}

	err := db.Get(&client, sql, clientID)
	if err != nil {
		return client, &DBOpError{sql, err}
	}

	return client, nil
}

func (u UserService) GetUserGrantsForClient(userID, clientID string) (oauth2.UserGrants, *DBOpError) {
	db := u.DB
	sql := `SELECT * FROM user_grants WHERE user_id  = $1 and client_id =$2`
	userGrants := oauth2.UserGrants{}

	err := db.Get(&userGrants, sql, userID, clientID)
	if err != nil {
		return userGrants, &DBOpError{sql, err}
	}

	return userGrants, nil
}

func (u UserService) UpdatePermissions(userGrants oauth2.UserGrants) *DBOpError {
	// check to see if the scope is valid for resource
	// ["profile:read", "profile:write"]
	// split by comma
	// further parse scope token by colon to get resource name and operation. e.g. profile for resource name read for operation
	_, dberr := u.GetUserGrantsForClient(userGrants.UserID, userGrants.ClientID)
	// if not found insert record
	if dberr != nil {
		sql := `INSERT INTO user_grants (user_id, client_id, scope) VALUES (:user_id, :client_id, :scope)`
		return u.NamedExec(sql, userGrants)
	} else {
		sql := `UPDATE user_grants SET scope = :scope WHERE user_id = :user_id AND client_id = :client_id`
		return u.NamedExec(sql, userGrants)
	}
}

func (u UserService) StoreAuthorizationRequestCode(ac oauth2.AuthorizationCode) *DBOpError {
	sql := `INSERT INTO authorization_code (code, client_id, created_at, user_id) VALUES (:code, :client_id, :created_at, :user_id)`
	return u.NamedExec(sql, ac)
}

func (u UserService) GetAuthorizationRequestCode(code, clientID string) (oauth2.AuthorizationCode, *DBOpError) {
	db := u.DB
	sql := `SELECT * FROM authorization_code WHERE code = $1 and client_id =$2`
	ac := oauth2.AuthorizationCode{}

	err := db.Get(&ac, sql, code, clientID)
	if err != nil {
		return ac, &DBOpError{sql, err}
	}

	return ac, nil
}

func (u UserService) AuthenticateClientUsingSecret(clientID, clientSecret string) (bool, *DBOpError) {
	client, dberr := u.GetClient(clientID)
	if dberr != nil {
		return false, dberr
	}
	return client.Secret == clientSecret, nil
}

func (u UserService) NonceUsedBefore(clientID, userID, nonce string) bool {
	return false
}

func (u UserService) StoreAccessTokenForUser(accessToken oauth2.AccessToken) *DBOpError {
	sql := `INSERT INTO access_token (token_id, user_id, token) VALUES (:token_id, :user_id, :token)`
	return u.NamedExec(sql, accessToken)
	//_, dberr := u.GetUserGrantsForClient(userGrants.UserID, userGrants.ClientID)
	//// if not found insert record
	//if dberr != nil {
	//	sql := `INSERT INTO user_grants (user_id, client_id, scope) VALUES (:user_id, :client_id, :scope)`
	//	return u.NamedExec(sql, userGrants)
	//} else {
	//	sql := `UPDATE user_grants SET scope = :scope WHERE user_id = :user_id AND client_id = :client_id`
	//	return u.NamedExec(sql, userGrants)
	//}
	return nil
}
