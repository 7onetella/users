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
    CONSTRAINT          unique_user UNIQUE (platform_name, email)
);
`

// User is application user
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
	user.MFAEnabled = false
	user.MFASecretTmpExp = 0
	user.MFASecretCurrent = ""

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

		dberr := service.Update(user)
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
