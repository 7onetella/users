package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mfcochauxlaberge/jsonapi"
	"log"
	"net/http"
)

type SecurityError struct {
	Event string
	Err   error
}

func (e *SecurityError) Unwrap() error {
	return e.Err
}

var dbschema = `
DROP TABLE users;

CREATE TABLE users (
    user_id             UUID PRIMARY KEY,
    platform_name       CHARACTER VARYING(64),
    email               CHARACTER VARYING(128) NOT NULL,
    passhash            CHARACTER VARYING(128) NOT NULL,
    firstname           CHARACTER VARYING(64),
    lastname            CHARACTER VARYING(64),
    created_date        BIGINT DEFAULT 0,
    mfa_enabled         BOOL DEFAULT FALSE,
    mfa_secret_current  CHARACTER VARYING(32) DEFAULT '',
    mfa_secret_tmp      CHARACTER VARYING(32) DEFAULT '',
    mfa_secret_tmp_exp  BIGINT DEFAULT 0,
	jwt_secret          CHARACTER VARYING(32) DEFAULT '',
    CONSTRAINT          unique_user UNIQUE (platform_name, email)
);

INSERT INTO users 
		(user_id, platform_name, email, passhash, firstname, lastname, created_date, mfa_enabled, mfa_secret_current, mfa_secret_tmp, mfa_secret_tmp_exp, jwt_secret) 
VALUES 
		('ee288e8c-0b2a-41b5-937c-9a355c0483b4', 'web', 'scott@example.com', 'password', 'scott', 'bar', 1597042574, true, 'FPTUDIF2KSQAKREU', '', 0, 'FPTUDIF2KSQAKREU');

INSERT INTO users 
		(user_id, platform_name, email, passhash, firstname, lastname, created_date, mfa_enabled, mfa_secret_current, mfa_secret_tmp, mfa_secret_tmp_exp, jwt_secret) 
VALUES 
		('a2aee5e6-05a0-438c-9276-4ba406b7bf9e', 'web', 'user8az28y@example.com', 'password', 'scott', 'bar', 1596747095, false, 'SVVEC5VTQBMNE3DH', 'C56BRBHMW3YC4XPA', 1597089055, 'SVVEC5VTQBMNE3DH');

DROP TABLE auth_event;

CREATE TABLE auth_event (
    event_id            UUID PRIMARY KEY,
    user_id             CHARACTER(40) DEFAULT '',
    event               CHARACTER VARYING(64) DEFAULT '',
    event_timestamp     BIGINT DEFAULT 0,
    ip_v4               CHARACTER(15) DEFAULT '',
    ip_v6               CHARACTER(38) DEFAULT '',
    agent               CHARACTER VARYING(128) DEFAULT ''
);
`

type AuthEvent struct {
	ID        string `db:"event_id"`
	UserIDReq string `db:"user_id"`
	Event     string `db:"event"`
	Timestamp int64  `db:"event_timestamp"`
	IPV4      string `db:"ip_v4"`
	IPV6      string `db:"ip_v6"`
	Agent     string `db:"agent"`
}

type User struct {
	ID               string `db:"user_id"       json:"id"            api:"users"`
	PlatformName     string `db:"platform_name" json:"platform_name" api:"attr"`
	Email            string `db:"email"         json:"email"         api:"attr"`
	Password         string `db:"passhash"      json:"password"      api:"attr"`
	FirstName        string `db:"firstname"     json:"firstname"     api:"attr"`
	LastName         string `db:"lastname"      json:"lastname"      api:"attr"`
	Created          int64  `db:"created_date"  json:"created"       api:"attr"`
	MFAEnabled       bool   `db:"mfa_enabled"   json:"mfaenabled"    api:"attr"`
	MFASecretCurrent string `db:"mfa_secret_current"`
	MFASecretTmp     string `db:"mfa_secret_tmp"`
	MFASecretTmpExp  int64  `db:"mfa_secret_tmp_exp"`
	JWTSecret        string `db:"jwt_secret"`
}

func UnmarshalUser(payload []byte, schema *jsonapi.Schema) (User, error) {
	doc, err := jsonapi.UnmarshalDocument([]byte(payload), schema)
	if err != nil {
		return User{}, err
	}
	res, ok := doc.Data.(jsonapi.Resource)
	if !ok {
		log.Println("type assertion")
		return User{}, errors.New("error while type asserting json resource")
	}
	user := User{}
	ID := res.Get("id").(string)
	if len(ID) > 0 {
		user.ID = ID
	}
	user.Email = res.Get("email").(string)
	user.Password = res.Get("password").(string)
	user.FirstName = res.Get("firstname").(string)
	user.LastName = res.Get("lastname").(string)
	user.MFAEnabled = res.Get("mfaenabled").(bool)

	return user, nil
}

func MarshalUser(uri string, schema *jsonapi.Schema, user User) (string, error) {
	//user.ID = ID
	doc := NewJSONDoc(user)
	return MarshalDoc(uri, schema, doc)
}

func ListUsers(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		users, dberr := service.List()
		if rh.HandleDBError(dberr) {
			return
		}

		doc := NewCollectionDoc(users)

		rh.SetContentTypeJSON()
		out, err := MarshalDoc(c.Request.URL.RequestURI(), userSchema, doc)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, out)
	}
}

func GetUser(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user granted access to his/her own account
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		id := c.Param("id")
		serr := rh.CheckUserIDMatchUserFromContext(id)
		if rh.HandleSecurityError(serr) {
			return
		}

		user, dberr := service.Get(id)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := MarshalUser(c.Request.URL.RequestURI(), userSchema, user)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, out)
	}
}

func DeleteUser(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		id := c.Param("id")
		serr := rh.CheckUserIDMatchUserFromContext(id)
		if rh.HandleSecurityError(serr) {
			return
		}

		dberr := service.Delete(id)
		if rh.HandleDBError(dberr) {
			return
		}

		c.Status(200)
	}
}

func UpdateUser(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		id := c.Param("id")
		serr := rh.CheckUserIDMatchUserFromContext(id)
		if rh.HandleSecurityError(serr) {
			return
		}

		payload, errs := rh.GetPayload(User{})
		if rh.HandleError(errs...) {
			return
		}

		user, err := UnmarshalUser(payload, userSchema)
		if rh.HandleError(err) {
			return
		}

		dberr := service.UpdateProfile(user)
		if rh.HandleDBError(dberr) {
			return
		}

		c.JSON(200, gin.H{
			"meta": gin.H{
				"result": "successful",
			},
		})
	}
}
