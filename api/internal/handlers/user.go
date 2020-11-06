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

// swagger:operation POST /users signup
//
//
// User signup using JSON:API document format as input
// ---
// summary: "Signup User"
// tags:
//   - account
// parameters:
//   - in: "body"
//     name: "body"
//     description: "User JSON:API Document"
//     required: true
//     schema:
//       "$ref": "#/definitions/JSONAPIUserSignup"
// produces:
//   - application/json
// responses:
//   '200':
//     description: user created
//     schema:
//       type: object
//       "$ref": "#/definitions/JSONAPIUserSignupResponse"
// security: []
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

func (ah AuthEventHandler) Handle(err error, code int, userID string, category Category, reason Reason) bool {
	if err != nil {
		rh := ah.RequestHanlder
		c := rh.Context
		userService := ah.UserService

		e := New(category, reason)
		c.AbortWithStatusJSON(code, e)

		event := rh.NewAuthEvent(userID, e.Message)

		dberr := userService.RecordAuthEvent(event)
		if dberr != nil {
			LogErr(rh.TransactionIDFromContext(), "db error", dberr.Err)
		}
		return true
	}
	return false
}

func (ah AuthEventHandler) Context() *gin.Context {
	return ah.RequestHanlder.Context
}

func (ah AuthEventHandler) DenyAccessForAnonymous(category Category, reason Reason) {
	c := ah.Context()
	e := New(category, reason)
	c.AbortWithStatusJSON(401, e)
	ah.RecordEvent("", e.ErrorCodeHuman)
}

func (ah AuthEventHandler) DenyAccessForUser(userID string, category Category, reason Reason) {
	c := ah.Context()
	e := New(category, reason)
	c.AbortWithStatusJSON(401, e)
	ah.RecordEvent(userID, e.ErrorCodeHuman)
}

func (ah AuthEventHandler) ExtractEventID(s string) (string, error) {
	decoded, err := crypto.Base64Decode(s)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func (ah AuthEventHandler) IsSigninSessionStillValid(timestamp int64, limit time.Duration) bool {
	return (time.Now().Unix() - timestamp) > int64(limit.Seconds())
}

func (ah AuthEventHandler) RecordEvent(userID, eventName string) {
	rh := ah.RequestHanlder
	event := rh.NewAuthEvent(userID, eventName)
	dberr := ah.UserService.RecordAuthEvent(event)
	if dberr != nil {
		LogErr(rh.TransactionIDFromContext(), "db error", dberr.Err)
	}
}

func (ah AuthEventHandler) AccessDeniedMissingData(userID string, category Category, reason Reason) {
	rh := ah.RequestHanlder
	c := rh.Context
	userService := ah.UserService

	e := New(category, reason)
	event := rh.NewAuthEvent(userID, e.ErrorCodeHuman)
	userService.RecordAuthEvent(event)

	mde := MissingDataError{
		Error:              *e,
		SigninSessionToken: crypto.Base64Encode(event.ID),
	}

	c.AbortWithStatusJSON(422, mde)
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
		"reason":                 event.Event,
		"message":                message,
		"signin_session_token":   crypto.Base64Encode(event.ID),
		"webauthn_session_token": crypto.Base64Encode(secEvent.ID),
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

// Reason is the last 3 digits of the error code.
type AuthType int

const (
	// Success indicates no error occurred.
	PasswordAuth AuthType = iota
	TOTPAuth
	WebauthnAuth
	UnknownAuth
)

func getAuthType(c Credentials) AuthType {
	if c.IsUsernamePresent() {
		return PasswordAuth
	}
	if c.IsTOTPCodePresent() {
		return TOTPAuth
	}
	if c.IsWebauthnTokenPresent() {
		return WebauthnAuth
	}
	return UnknownAuth
}

// swagger:operation POST /jwt_auth/signin signin
//
// Signs in user with MFA support.
//
// Signs in user with optional MFA support. TOTP and WebAuthn can be enabled to add stronger method of authentication.
// ---
// summary: "Signin User"
// tags:
//   - account
// parameters:
//   - in: "body"
//     name: "body"
//     description: "User Credentials"
//     required: true
//     schema:
//       "$ref": "#/definitions/CredentialsBase"
// produces:
//   - application/json
// responses:
//   '200':
//     description: access granted
//     schema:
//       type: object
//       "$ref": "#/definitions/AuthToken"
//   '401':
//     description: access denied
//     schema:
//       type: object
//       "$ref": "#/definitions/AuthError"
//   '422':
//     description: missing data
//     schema:
//       type: object
//       "$ref": "#/definitions/MissingDataError"
// security: []
func Signin(userService UserService, claimKey string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		auth := AuthEventHandler{rh, userService}

		rh.WriteCORSHeader()

		cred := Credentials{}
		err := c.ShouldBind(&cred)
		if err != nil {
			auth.DenyAccessForAnonymous(JSONError, Unmarshalling)
			return
		}

		log.Printf("cred = %#v", cred)

		var user User

		//switch getAuthType(cred) {
		//case PasswordAuth:
		//	break
		//case TOTPAuth:
		//
		//}

		// if 2fa and password signin was already completed
		if cred.IsSigninSessionTokenPresent() {
			eventID, err := auth.ExtractEventID(cred.SigninSessionToken)
			if err != nil {
				auth.DenyAccessForAnonymous(AuthenticationError, SigninSessionTokenDecodingFailed)
				return
			}
			e, dberr := userService.GetAuthEvent(eventID)
			if dberr != nil {
				auth.DenyAccessForAnonymous(DatabaseError, QueryingFailed)
				return
			}
			if auth.IsSigninSessionStillValid(e.Timestamp, time.Minute*5) {
				auth.DenyAccessForUser(e.UserID, AuthenticationError, SigninSessionExpired)
				return
			}
			userFromDB, dberr := userService.Get(e.UserID)
			if dberr != nil {
				auth.DenyAccessForUser(e.UserID, DatabaseError, QueryingFailed)
				return
			}
			user = userFromDB
			goto CheckIf2FARequired
		}

		if cred.IsUsernamePresent() {
			userFromDB, dberr := auth.UserService.FindByEmail(cred.Username)
			if dberr != nil {
				auth.DenyAccessForAnonymous(DatabaseError, QueryingFailed)
				return
			}

			if userFromDB.Password != cred.Password {
				auth.DenyAccessForUser(user.ID, AuthenticationError, UsernameOrPasswordDoesNotMatch)
				return
			}
			user = userFromDB
		}

	CheckIf2FARequired:
		if user.WebAuthnEnabled {
			if !cred.IsWebauthnTokenPresent() {
				auth.AccessDeniedMissingData(user.ID, AuthenticationError, WebAuthnRequired)
				return
			}
			if !auth.IsWebAuthnAuthTokenValidForUer(user.ID, cred.WebauthnAuthToken) {
				auth.DenyAccessForUser(user.ID, AuthenticationError, WebauthnAuthFailure)
				return
			}
			goto GrantAccess
		}

		if user.TOTPEnabled {
			if !cred.IsTOTPCodePresent() {
				auth.AccessDeniedMissingData(user.ID, AuthenticationError, TOTPRequired)
				return
			}
			if !IsTOTPValid(user, cred.TOTP) {
				auth.DenyAccessForUser(user.ID, AuthenticationError, TOTPAuthFailure)
				return
			}
			goto GrantAccess
		}

	GrantAccess:
		// guard against empty json payload
		if len(user.ID) == 0 {
			log.Printf("user = %#v", user)
			auth.DenyAccessForAnonymous(AuthenticationError, UserUnknown)
			return
		}

		tokenString, expTime, err := EncodeToken(user.ID, user.JWTSecret, ttl)
		if err != nil {
			log.Println("encoding error")
			auth.DenyAccessForUser(user.ID, AuthenticationError, JWTEncodingFailure)
			return
		}

		token := AuthToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}
		c.JSON(200, token)

		auth.RecordEvent(user.ID, "login_successful")
	}
}

// swagger:operation GET /users/{id} getuser
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
//       "$ref": "#/definitions/JSONAPIUserSignupResponse"
// security:
//   - bearer_token: []
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

// swagger:operation DELETE /users/{id} deleteuser
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
//   - bearer_token: []
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

// swagger:operation PATCH /users/{id} updateuser
//
// ---
// summary: "Updates user profile"
// tags:
//   - profile
// parameters:
//   - in: "body"
//     name: "body"
//     description: "User JSON:API Document"
//     required: true
//     schema:
//       "$ref": "#/definitions/JSONAPIUserSignup"
// produces:
//   - application/json
// responses:
//   '204':
//     description: user updated
// security:
//   - bearer_token: []
//   - oauth2:
//	     - 'write:profile'
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
