package handlers

import (
	"errors"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/jsonutil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/crypto"
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

type AuthEventHandler struct {
	RequestHanlder RequestHanlder
	UserService    UserService
}

func (ah AuthEventHandler) ErrorOccurred(eventName, message string) {
	rh := ah.RequestHanlder
	c := rh.Context
	userService := ah.UserService

	event := rh.NewAuthEvent("", eventName)
	c.AbortWithStatusJSON(401, gin.H{
		"reason":  event.Event,
		"message": message,
	})
	dberr := userService.RecordAuthEvent(event)
	if dberr != nil {
		LogErr(rh.TransactionIDFromContext(), "db error", dberr.Err)
	}
}

func (ah AuthEventHandler) AccessDenied(userID, eventName, message string) {
	rh := ah.RequestHanlder
	c := rh.Context
	userService := ah.UserService

	event := rh.NewAuthEvent(userID, eventName)
	c.AbortWithStatusJSON(401, gin.H{
		"reason":  event.Event,
		"message": message,
	})
	dberr := userService.RecordAuthEvent(event)
	if dberr != nil {
		LogErr(rh.TransactionIDFromContext(), "db error", dberr.Err)
	}
}

func (ah AuthEventHandler) SendPrimaryAuthToken(userID, eventName, message string) {
	rh := ah.RequestHanlder
	c := rh.Context
	userService := ah.UserService

	event := rh.NewAuthEvent(userID, eventName)
	userService.RecordAuthEvent(event)

	c.AbortWithStatusJSON(401, gin.H{
		"reason":     event.Event,
		"message":    message,
		"auth_token": crypto.Base64Encode(event.ID),
	})
}

func (ah AuthEventHandler) FinishSecondAuth(userID, eventName, message string) {
	rh := ah.RequestHanlder
	c := rh.Context
	userService := ah.UserService

	event := rh.NewAuthEvent(userID, eventName)
	userService.RecordAuthEvent(event)

	secEvent := rh.NewAuthEvent(userID, "sec_auth_generated")
	userService.RecordAuthEvent(secEvent)

	c.AbortWithStatusJSON(401, gin.H{
		"reason":         event.Event,
		"message":        message,
		"auth_token":     crypto.Base64Encode(event.ID),
		"sec_auth_token": crypto.Base64Encode(secEvent.ID),
	})
}

func (ah AuthEventHandler) IsSecondaryAuthTokenValidForUer(userID, secAuthToken string) bool {
	userService := ah.UserService
	eventID, _ := crypto.Base64Decode(secAuthToken)
	user, dberr := userService.FindUserByAuthEventID(eventID)
	if dberr != nil {
		return false
	}

	return userID == user.ID
}

// Signin signs in user
func Signin(userService UserService, claimKey string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		auth := AuthEventHandler{rh, userService}

		rh.WriteCORSHeader()

		cred := Credentials{}
		c.ShouldBind(&cred)

		var user User
		var dberr *DBOpError

		if len(cred.PrimaryAuthToken) > 0 {
			eventID, _ := crypto.Base64Decode(cred.PrimaryAuthToken)
			e, _ := userService.GetAuthEvent(eventID)
			// if it took more than 5 mins after initial pass login
			if (time.Now().Unix() - e.Timestamp) > 300 {
				auth.AccessDenied(e.UserID, "login_auth_expired", "Your login session timed out")
				return
			}
			user, dberr = userService.Get(e.UserID)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				auth.ErrorOccurred("invalid_auth_token", "Authentication failed")
				return
			}
			goto Check2FA
		}

		if len(cred.Username) > 0 {
			user, dberr = userService.FindByEmail(cred.Username)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				auth.ErrorOccurred("server_error", "Authentication failed")
				return
			}

			if user.Password != cred.Password {
				auth.AccessDenied(user.ID, "login_password_invalid", "Check login name or password")
				return
			}
		}

	Check2FA:
		if user.WebAuthnEnabled {
			if len(cred.SecondaryAuthToken) == 0 {
				auth.SendPrimaryAuthToken(user.ID, "login_webauthn_requested", "WebAuthn Auth Required")
				return
			}
			if !auth.IsSecondaryAuthTokenValidForUer(user.ID, cred.SecondaryAuthToken) {
				auth.AccessDenied(user.ID, "invalid_sec_auth_token", "Authentication failed")
				return
			}
			goto GrantAccess
		}

		if user.TOTPEnabled {
			if len(cred.TOTP) == 0 {
				auth.SendPrimaryAuthToken(user.ID, "login_totp_requested", "TOTP required")
				return
			}
			if !IsTOTPValid(user, cred.TOTP) {
				auth.AccessDenied(user.ID, "login_totp_invalid", "Your code is invalid")
				return
			}
			goto GrantAccess
		}

	GrantAccess:
		tokenString, expTime, err := jwtutil.EncodeToken(claimKey, user.ID, user.JWTSecret, ttl)
		if err != nil {
			log.Println("encoding error")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		log.Println("Signin successful")

		token := AuthToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}
		c.JSON(200, token)

		userService.RecordAuthEvent(NewAuthEvent(user.ID, "login_successful", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
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
