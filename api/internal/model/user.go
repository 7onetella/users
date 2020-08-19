package model

import "github.com/duo-labs/webauthn/webauthn"

type User struct {
	ID                  string `db:"user_id"              json:"id"                 api:"users"`
	PlatformName        string `db:"platform_name"        json:"platform_name"      api:"attr"`
	Email               string `db:"email"                json:"email"              api:"attr"`
	Password            string `db:"passhash"             json:"password"           api:"attr"`
	FirstName           string `db:"firstname"            json:"firstname"          api:"attr"`
	LastName            string `db:"lastname"             json:"lastname"           api:"attr"`
	Created             int64  `db:"created_date"         json:"created"            api:"attr"`
	TOTPEnabled         bool   `db:"totp_enabled"         json:"totpenabled"        api:"attr"`
	TOTPSecretCurrent   string `db:"totp_secret_current"`
	TOTPSecretTmp       string `db:"totp_secret_tmp"`
	TOTPSecretTmpExp    int64  `db:"totp_secret_tmp_exp"`
	WebAuthnEnabled     bool   `db:"webauthn_enabled"     json:"webauthnenabled"    api:"attr"`
	WebAuthnSessionData string `db:"webauthn_session"`
	JWTSecret           string `db:"jwt_secret"`
	Credentials         []webauthn.Credential
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
