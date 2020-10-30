package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/jsonutil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/crypto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mfcochauxlaberge/jsonapi"
	"github.com/xlzd/gotp"
)

var UserJSONSchema *jsonapi.Schema

func init() {
	s, errs := SchemaCheck(&User{})
	if errs != nil && len(errs) > 0 {
		log.Fatalln(errs)
	}
	UserJSONSchema = s
}

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

func (ah AuthEventHandler) IsWebAuthnAuthTokenValidForUer(userID, webauthnAuthToken string) bool {
	userService := ah.UserService
	eventID, _ := crypto.Base64Decode(webauthnAuthToken)
	user, dberr := userService.FindUserByAuthEventID(eventID)
	if dberr != nil {
		return false
	}

	return userID == user.ID
}

// swagger:operation POST /jwt_auth/signin signin
//
// ---
// summary: "Signin user"
// tags:
//   - signin
// parameters:
//   - in: "body"
//     name: "body"
//     description: "User credentials"
//     required: true
//     schema:
//       "$ref": "#/definitions/Credentials"
// produces:
//   - application/json
// responses:
//   '200':
//     description: delete user profile response
// security:
//   - api_key: []
func Signin(userService UserService, claimKey string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		auth := AuthEventHandler{rh, userService}

		rh.WriteCORSHeader()

		cred := Credentials{}
		err := c.ShouldBind(&cred)
		if err != nil {
			log.Printf("err = %v", err)
			auth.AccessDenied("", "binding_error", "Binding Error")
			return
		}

		log.Printf("cred = %#v", cred)

		var user User
		var dberr *DBOpError

		// if 2fa and password signin was already completed
		if len(cred.SigninSessionToken) > 0 {
			eventID, err := crypto.Base64Decode(cred.SigninSessionToken)
			log.Printf("event id = %s", eventID)
			if err != nil {
				auth.AccessDenied("", "error_decoding_auth_token", "Error decoding auth_token")
				return
			}
			e, dberr := userService.GetAuthEvent(eventID)
			if dberr != nil {
				log.Printf("error while getting auth event: %v", dberr)
				auth.ErrorOccurred("err_get_auth_event", "Error getting auth event")
				return
			}
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
			if len(cred.WebauthnAuthToken) == 0 {
				auth.SendPrimaryAuthToken(user.ID, "login_webauthn_requested", "WebAuthn Auth Required")
				return
			}
			if !auth.IsWebAuthnAuthTokenValidForUer(user.ID, cred.WebauthnAuthToken) {
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
		// safe guard for empty json payload
		if len(user.ID) == 0 {
			log.Printf("user = %#v", user)
			auth.AccessDenied(user.ID, "user_id_invalid", "invalid user")
			return
		}

		tokenString, expTime, err := EncodeToken(user.ID, user.JWTSecret, ttl)
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

// swagger:operation GET /users/{id} profile
//
// ---
// summary: "Get user profile"
// tags:
//   - profile
// produces:
//   - application/json
// responses:
//   '200':
//     description: get user profile response
//     schema:
//       "$ref": "#/definitions/JSONAPIUser"
// security:
//   - api_key: []
//   - oauth2:
//	     - 'read:profile'
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

// swagger:operation DELETE /users/{id} profile
//
// ---
// summary: "Delete user profile"
// tags:
//   - profile
// produces:
//   - application/json
// responses:
//   '200':
//     description: delete user profile response
// security:
//   - api_key: []
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

		c.Status(204)
		//c.JSON(200, gin.H{
		//	"meta": gin.H{
		//		"result": "successful",
		//	},
		//})
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
