package model

import "github.com/duo-labs/webauthn/webauthn"

// swagger:model User
type User struct {
	// the id for this user
	//
	// required: true
	// read-only: true
	// example: a2aee5e6-05a0-438c-9276-4ba406b7bf9e
	ID           string `json:"id"              db:"user_id"               api:"users"`
	PlatformName string `json:"-"               db:"platform_name"`
	// email
	//
	// required: true
	// example: john.smith@example.com
	Email string `json:"email"           db:"email"                 api:"attr"`
	// password
	//
	// required: true
	// example: pass1234
	Password string `json:"password"        db:"passhash"              api:"attr"`
	// first name
	//
	// required: true
	// example: John
	FirstName string `json:"firstname"       db:"firstname"             api:"attr"`
	// last name
	//
	// required: true
	// example: Smith
	LastName string `json:"lastname"        db:"lastname"              api:"attr"`
	// created date in unix time
	//
	// read-only: true
	// example: 1596747095
	Created int64 `json:"created"           db:"created_date"          api:"attr"`
	// totp enabled or not
	//
	// example: false
	TOTPEnabled       bool   `json:"totpenabled"       db:"totp_enabled"          api:"attr"`
	TOTPSecretCurrent string `json:"-"                 db:"totp_secret_current"`
	TOTPSecretTmp     string `json:"-"                 db:"totp_secret_tmp"`
	TOTPSecretTmpExp  int64  `json:"-"                 db:"totp_secret_tmp_exp"`
	// webauthn enabled or not
	//
	// example: false
	WebAuthnEnabled     bool                  `json:"webauthnenabled"   db:"webauthn_enabled"      api:"attr"`
	WebAuthnSessionData string                `json:"-"                 db:"webauthn_session"`
	JWTSecret           string                `json:"-"                 db:"jwt_secret"`
	Credentials         []webauthn.Credential `json:"-"`
}

// UserProfile represents the user for this application
//
// A user profile is very narrow representation of user
//
// swagger:model UserProfile
type UserProfile struct {
	// the id for this user
	//
	// read-only: true
	// example: a2aee5e6-05a0-438c-9276-4ba406b7bf9e
	ID string `json:"id"              db:"user_id"               api:"users"`
	// the email
	//
	// read-only: true
	// example: john.smith@example.com
	Email string `json:"email"           db:"email"                 api:"attr"`
	// the first name
	//
	// read-only: true
	// example: John
	FirstName string `json:"firstname"       db:"firstname"             api:"attr"`
	// the last name
	//
	// read-only: true
	// example: Smith
	LastName string `json:"lastname"        db:"lastname"              api:"attr"`
}

func (u User) WebAuthnID() []byte {
	return []byte(u.ID)
}

func (u User) WebAuthnName() string {
	return u.Email
}

func (u User) WebAuthnDisplayName() string {
	return u.FirstName + " " + u.LastName
}

func (u User) WebAuthnIcon() string {
	return ""
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u User) AddCredential(credential webauthn.Credential) {
	u.Credentials = append(u.Credentials, credential)
}
