package model

import "github.com/duo-labs/webauthn/webauthn"

type User struct {
	ID                  string                `json:"id"              example:"a2aee5e6-05a0-438c-9276-4ba406b7bf9e"    db:"user_id"               api:"users"`
	PlatformName        string                `json:"platform_name"   example:"web"                                     db:"platform_name"         api:"attr"`
	Email               string                `json:"email"           example:"foo.bar@example.com"                     db:"email"                 api:"attr"`
	Password            string                `json:"password"        example:"users91234"                              db:"passhash"              api:"attr"`
	FirstName           string                `json:"firstname"       example:"foo"                                     db:"firstname"             api:"attr"`
	LastName            string                `json:"lastname"        example:"bar"                                     db:"lastname"              api:"attr"`
	Created             int64                 `json:"created"         example:"1596747095"                              db:"created_date"          api:"attr"`
	TOTPEnabled         bool                  `json:"totpenabled"     example:"false"                                   db:"totp_enabled"          api:"attr"`
	TOTPSecretCurrent   string                `json:"-"               example:""                                        db:"totp_secret_current"`
	TOTPSecretTmp       string                `json:"-"               example:""                                        db:"totp_secret_tmp"`
	TOTPSecretTmpExp    int64                 `json:"-"               example:""                                        db:"totp_secret_tmp_exp"`
	WebAuthnEnabled     bool                  `json:"webauthnenabled" example:"false"                                   db:"webauthn_enabled"      api:"attr"`
	WebAuthnSessionData string                `json:"-"               example:""                                        db:"webauthn_session"`
	JWTSecret           string                `json:"-"               example:""                                        db:"jwt_secret"`
	Credentials         []webauthn.Credential `json:"-"`
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
