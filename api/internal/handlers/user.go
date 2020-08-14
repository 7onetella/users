package handlers

import (
	"errors"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/jsonutil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/mfcochauxlaberge/jsonapi"
	"log"
	"net/http"
)

var UserJSONSchema *jsonapi.Schema

func init() {
	s, errs := SchemaCheck(&User{})
	if errs != nil && len(errs) > 0 {
		log.Fatalln(errs)
	}
	UserJSONSchema = s
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
		out, err := MarshalUser(c.Request.URL.RequestURI(), UserJSONSchema, user)
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

		user, err := UnmarshalUser(payload, UserJSONSchema)
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
	user.TOTPEnabled = res.Get("totpenabled").(bool)

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
		out, err := MarshalDoc(c.Request.URL.RequestURI(), UserJSONSchema, doc)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, out)
	}
}
