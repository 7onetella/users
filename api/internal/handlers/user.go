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
// ---
// summary: Signup
// description: |
//   User signup using JSON:API document format as input
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

		//rh.Logf("user.signup.info payload=%s", string(payload))
		user, err := UnmarshalUser(payload, UserJSONSchema)
		if err != nil {
			rh.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling, err)
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
		if err != nil {
			rh.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling, err)
			return
		}
		c.String(http.StatusOK, out)
	}
}

// swagger:operation POST /signin signin
//
// ---
// summary: Signin
// description: |
//
//   Here is the list of authentication types
//
//   |      1st Factor             |          2nd Factor       |     Security                      |
//   |-----------------------------|---------------------------|-----------------------------------|
//   |      Password               |             None          | Weak - password can be guessed    |
//   |      Password               |             TOTP          | Medium - prone to phishing attack |
//   |      Password               |             WebAuthn      | Strong                            |
//
//   **Password Only**
//
//   Successful password authentication call to this endpoint sends back http status code `200` with JWT auth token.
//
//   **Password + TOTP**
//
//   This is two factor authentication. User's security settings must have TOTP enabled in order for this second
//   factor authentication to be presented to the user after a successful password authentication.
//   Successful password auth call to this endpoint results in the server sending back http status code `422` with
//   `signin_session_token` which represents successful password authentication. The client at that point is
//   expected to send TOTP with received `signin_session_token`. This endpoint will validate both data. Once valid the
//   endpoint will send back http status code `200` with JWT auth token.
//
//   <img src="/accounts/assets/totp_flow.png" alt="TOTP Flow">
//
//   **Password + WebAuthn**
//
//   This is two factor authentication. User's security settings must have WebAuthn enabled in order for this second
//   factor authentication to be presented to the user after a successful password authentication. Calling WebAuthn
//   endpoint requires working with browser's <a href="https://developer.mozilla.org/en-US/docs/Web/API/Web_Authentication_API">**WebAuthn API**</a>
//   That brings a bit of challenge to automated testing. Testing tool can not simply pass in some value to this restful
//   service endpoint. The tool needs to call browser's WebAuthn API to have the browser initiate a communication
//   with the user's physical authenticator device. Although it is theoretically possible to emulate the physical authenticator
//   using <a href="https://github.com/github/SoftU2F/">**software**</a>, there isn't a software authenticator that is suitable
//   for fully automated testing. Therefore, testing of WebAuthn flow will remain as manual.
//
//   <img src="/accounts/assets/webauthn_flow.png" alt="WebAuthn Flow">
//
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
//       "$ref": "#/definitions/JWTToken"
//   '401':
//     description: access denied
//     schema:
//       type: object
//       properties:
//         code:
//           type: integer
//           description: error code
//           example: 4300
//         message:
//           type: string
//           description: error message
//           example: Check login name or password
//   '422':
//     description: missing data
//     schema:
//       type: object
//       properties:
//         code:
//           type: integer
//           description: error code
//           example: 4800
//         message:
//           type: string
//           description: error message
//           example: TOTP required
//         signin_session_token:
//           type: string
//           description: signin session token
//           example: MzM4OGNkMWEtNmQyNC00MDQ1LWJmYzctMWJlMzM3ZTk1NDQ5
// security: []
// x-codeSamples:
//   - lang: curl-password
//     source: |
//       curl -X POST \
//       https://accounts.7onetella.net/signin \
//       -H 'Content-Type: application/json' \
//       -d '{
//	       "username": "john.smith@example.com",
//	       "password": "password1234"
//       }'
//   - lang: curl-totp
//     source: |
//       curl -X POST \
//       https://accounts.7onetella.net/signin \
//       -H 'Content-Type: application/json' \
//       -d '{
//         "signin_session_token": "YjcxYmE0MjQtMGY1MS00ZDI0LTk4NDAtYjhiM2IwNzY5ZDNk",
//         "totp": "592918"
//       }'
func Signin(userService UserService, ttl time.Duration, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		auth := AuthEventHandler{rh, userService}

		rh.WriteCORSHeader()

		cred := Credentials{}
		err := c.ShouldBind(&cred)
		if err != nil {
			auth.DenyAccessForAnonymous(New(JSONAPISpecError, Unmarshalling))
			return
		}

		var user User

		// if 2fa and password signin was already completed
		if cred.IsSigninSessionTokenPresent() {
			eventID, err := auth.ExtractEventID(cred.SigninSessionToken)
			if err != nil {
				rh.Logf("signin_session_token=%s", cred.SigninSessionToken)
				auth.DenyAccessForAnonymous(New(AuthenticationError, SigninSessionTokenDecodingFailed))
				return
			}
			e, dberr := userService.GetAuthEvent(eventID)
			if dberr != nil {
				rh.Logf("querying for event_id %s", eventID)
				auth.DenyAccessForAnonymous(New(DatabaseError, QueryingFailed))
				return
			}
			if auth.IsSigninSessionStillValid(e.Timestamp, time.Minute*5) {
				auth.DenyAccessForUser(e.UserID, New(AuthenticationError, SigninSessionExpired))
				return
			}
			userFromDB, dberr := userService.Get(e.UserID)
			if dberr != nil {
				auth.DenyAccessForUser(e.UserID, Wrap(DatabaseError, QueryingFailed, err))
				return
			}
			user = userFromDB
			goto CheckIf2FARequired
		}

		if cred.IsUsernamePresent() {
			userFromDB, dberr := userService.FindByEmail(cred.Username)
			if dberr != nil {
				auth.DenyAccessForAnonymous(Wrap(DatabaseError, QueryingFailed, dberr))
				return
			}

			if userFromDB.Password != cred.Password {
				auth.DenyAccessForUser(user.ID, New(AuthenticationError, UsernameOrPasswordDoesNotMatch))
				return
			}
			user = userFromDB
		}

	CheckIf2FARequired:
		if user.WebAuthnEnabled {
			if !cred.IsWebauthnTokenPresent() {
				auth.AccessDeniedMissingData(user.ID, New(AuthenticationError, WebAuthnRequired))
				return
			}
			if !auth.IsWebAuthnSessionTokenValidForUer(user.ID, cred.WebAuthnSessionToken) {
				auth.DenyAccessForUser(user.ID, New(AuthenticationError, WebauthnAuthFailure))
				return
			}
			goto GrantAccess
		}

		if user.TOTPEnabled {
			if !cred.IsTOTPCodePresent() {
				auth.AccessDeniedMissingData(user.ID, New(AuthenticationError, TOTPRequired))
				return
			}
			if !IsTOTPValid(user, cred.TOTP) {
				auth.DenyAccessForUser(user.ID, New(AuthenticationError, TOTPAuthFailure))
				return
			}
			goto GrantAccess
		}

	GrantAccess:
		// guard against empty json payload
		if len(user.ID) == 0 {
			auth.DenyAccessForAnonymous(New(AuthenticationError, UserUnknown))
			return
		}

		tokenString, expTime, err := EncodeToken(user.ID, user.JWTSecret, issuer, ttl)
		if err != nil {
			auth.DenyAccessForUser(user.ID, Wrap(AuthenticationError, JWTEncodingFailure, err))
			return
		}

		token := JWTToken{
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
//     description: user profile
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
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, rh.WrapAsJSONAPIErrors(serr))
			return
		}

		user, dberr := service.Get(id)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := MarshalUser(c.Request.URL.RequestURI(), UserJSONSchema, user)
		if err != nil {
			LogErr(rh.TX(), "error marshalling user", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Marshalling))
			return
		}

		c.String(http.StatusOK, out)
	}
}

// swagger:operation DELETE /users/{id} deleteuser
//
// ---
// summary: Delete user profile
// tags:
//   - profile
// produces:
//   - application/json
// responses:
//   '200':
//     description: user profile deleted
// security:
//   - bearer_token: []
func DeleteUser(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		id := c.Param("id")
		serr := rh.CheckUserIDMatchUserFromContext(id)
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, rh.WrapAsJSONAPIErrors(serr))
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
// summary: Updates user profile
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
//     description: user profile updated
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
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, rh.WrapAsJSONAPIErrors(serr))
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
	doc, err := jsonapi.UnmarshalDocument(payload, schema)
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
