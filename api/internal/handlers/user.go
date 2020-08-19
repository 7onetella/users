package handlers

import (
	"errors"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/jsonutil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mfcochauxlaberge/jsonapi"
	"github.com/xlzd/gotp"
	"log"
	"net/http"
	"time"
)

var UserJSONSchema *jsonapi.Schema

func init() {
	s, errs := SchemaCheck(&User{})
	if errs != nil && len(errs) > 0 {
		log.Fatalln(errs)
	}
	UserJSONSchema = s
}

// Signup signs up user
func Signup(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)

		rh.WriteCORSHeader()

		payload, errs := rh.GetPayload(User{})
		if rh.HandleError(errs...) {
			return
		}

		user, err := UnmarshalUser(payload, UserJSONSchema)
		if rh.HandleError(err) {
			return
		}

		user.ID = uuid.New().String()
		user.Created = CurrentTimeInSeconds()
		user.PlatformName = "web"
		user.JWTSecret = gotp.RandomSecret(16)

		dberr := userService.Register(user)
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

// Signin signs in user
func Signin(userService UserService, claimKey string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		cred := Credentials{}
		c.ShouldBind(&cred)

		var user *User
		var dberr *DBOpError

		if len(cred.EventID) > 0 {
			user, dberr = userService.FindByEventID(cred.EventID)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_event_id",
					"message": "TOTP auth failed",
				})
				return
			}
			goto Check2FA
		}

		if len(cred.Username) > 0 {
			user, dberr = userService.FindByEmail(cred.Username)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "server_error",
					"message": "Authentication failed",
				})
			}

			if user.Password != cred.Password {
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_credential",
					"message": "Check login name or password",
				})
				dberr := userService.RecordAuthEvent(NewAuthEvent(user.ID, "invalid_credential", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
				if rh.HandleDBError(dberr) {
					return
				}
				return
			}

			// this just to make testing easier during development phase
			if user.Email == "foo_pass_user@example.com" {
				goto GrantAccess
			}
		}

	Check2FA:
		if user.TOTPEnabled {
			if len(cred.TOTP) == 0 {
				event := NewAuthEvent(user.ID, "missing_totp", c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
				userService.RecordAuthEvent(event)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":   "missing_totp",
					"message":  "TOTP required",
					"event_id": event.ID,
				})
				return
			}
			if !IsTOTPValid(user, cred.TOTP) {
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_totp",
					"message": "Check your TOTP",
				})
				userService.RecordAuthEvent(NewAuthEvent(user.ID, "invalid_totp", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
				return
			}
		}
		if user.WebAuthnEnabled {
			if len(cred.TokenID) == 0 {
				event := NewAuthEvent(user.ID, "webauthn_required", c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
				userService.RecordAuthEvent(event)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":   "webauthn_required",
					"message":  "WebAuthn Auth Required",
					"event_id": event.ID,
				})
				return
			}
			// TODO: make sure token id match
		}

	GrantAccess:
		tokenString, expTime, err := jwtutil.EncodeToken(claimKey, user.ID, user.JWTSecret, ttl)
		if err != nil {
			log.Println("encoding error")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		log.Println("Signin successful")
		//log.Println("Sign-In successful dropping token cookie")
		//http.SetCookie(w, &http.Cookie{
		//	Name:    "token",
		//	Value:   tokenString,
		//	Expires: expTime,
		//})

		token := AuthToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}
		c.JSON(200, token)

		userService.RecordAuthEvent(NewAuthEvent(user.ID, "successful_login", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
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
