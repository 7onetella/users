package model

type User struct {
	ID                string `db:"user_id"        json:"id"             api:"users"`
	PlatformName      string `db:"platform_name"  json:"platform_name"  api:"attr"`
	Email             string `db:"email"          json:"email"          api:"attr"`
	Password          string `db:"passhash"       json:"password"       api:"attr"`
	FirstName         string `db:"firstname"      json:"firstname"      api:"attr"`
	LastName          string `db:"lastname"       json:"lastname"       api:"attr"`
	Created           int64  `db:"created_date"   json:"created"        api:"attr"`
	TOTPEnabled       bool   `db:"totp_enabled"   json:"totpenabled"    api:"attr"`
	TOTPSecretCurrent string `db:"totp_secret_current"`
	TOTPSecretTmp     string `db:"totp_secret_tmp"`
	TOTPSecretTmpExp  int64  `db:"totp_secret_tmp_exp"`
	JWTSecret         string `db:"jwt_secret"`
}
