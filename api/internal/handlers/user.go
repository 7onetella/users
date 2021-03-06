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
//   '500':
//     description: internal server issue
// security: []
func Signup(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)

		r.WriteCORSHeader()

		payload, err := r.GetBody()
		if err != nil {
			r.AbortWithStatusInternalServerError(ServerError, RetrievingPayloadError)
			return
		}

		user, err := UnmarshalUser(payload, UserJSONSchema)
		if err != nil {
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
			return
		}

		user.ID = uuid.New().String()
		user.Created = CurrentTimeInSeconds()
		user.PlatformName = "web"
		user.JWTSecret = gotp.RandomSecret(16)

		dberr := userService.Register(user)
		if dberr != nil {
			r.Logf("signup.register.failed err=%s", dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, PersistingFailed)
		}

		r.SetContentTypeJSON()
		uri := c.Request.URL.RequestURI()
		response, err := MarshalUser(uri, UserJSONSchema, user)
		if err != nil {
			r.Logf("signup.marshall-user.failed uri=% schema=%s err=%s", uri, UserJSONSchema, err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}
		c.String(http.StatusOK, response)
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

		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		cred := Credentials{}
		err := c.ShouldBind(&cred)
		if err != nil {
			r.Logf("signin.jsong-binding.failed error=%s", err)
			r.AbortWithStatusUnauthorizedError(JSONAPISpecError, Unmarshalling)
			return
		}

		var user User

		// if 2fa and password signin was already completed
		if cred.IsSigninSessionTokenPresent() {
			eventID, err := r.ExtractEventID(cred.SigninSessionToken)
			if err != nil {
				r.Logf("signin.decode-userid.failed signin_session_token=%s error=%s", cred.SigninSessionToken, err)
				r.AbortWithStatusUnauthorizedError(AuthenticationError, SigninSessionTokenDecodingFailed)
				return
			}
			e, dberr := userService.GetAuthEvent(eventID)
			if dberr != nil {
				r.Logf("signin.get-auth-event.failed event_id=%s error=%v", eventID, dberr)
				r.AbortWithStatusUnauthorizedError(DatabaseError, QueryingFailed)
				return
			}
			if r.IsSigninSessionStillValid(e.Timestamp, time.Minute*5) {
				r.AbortWithStatusUnauthorizedError(AuthenticationError, SigninSessionExpired)
				return
			}
			userFromDB, dberr := userService.Get(e.UserID)
			if dberr != nil {
				r.Logf("querying for user_id %s", e.UserID)
				r.DenyAccessForUser(e.UserID, DatabaseError, QueryingFailed)
				return
			}
			user = userFromDB
			goto CheckIf2FARequired
		}

		if cred.IsUsernamePresent() {
			userFromDB, dberr := userService.FindByEmail(cred.Username)
			if dberr != nil {
				r.DenyAccessForAnonymous(DatabaseError, QueryingFailed)
				return
			}

			if userFromDB.Password != cred.Password {
				r.DenyAccessForUser(user.ID, AuthenticationError, UsernameOrPasswordDoesNotMatch)
				return
			}
			user = userFromDB
		}

	CheckIf2FARequired:
		if user.WebAuthnEnabled {
			if !cred.IsWebauthnTokenPresent() {
				r.AccessDeniedMissingData(user.ID, AuthenticationError, WebAuthnRequired)
				return
			}
			if !r.IsWebAuthnSessionTokenValidForUer(user.ID, cred.WebAuthnSessionToken) {
				r.DenyAccessForUser(user.ID, AuthenticationError, WebauthnAuthFailure)
				return
			}
			goto GrantAccess
		}

		if user.TOTPEnabled {
			if !cred.IsTOTPCodePresent() {
				r.AccessDeniedMissingData(user.ID, AuthenticationError, TOTPRequired)
				return
			}
			if !IsTOTPValid(user, cred.TOTP) {
				r.DenyAccessForUser(user.ID, AuthenticationError, TOTPAuthFailure)
				return
			}
			goto GrantAccess
		}

	GrantAccess:
		// guard against empty json payload
		if len(user.ID) == 0 {
			r.DenyAccessForAnonymous(AuthenticationError, UserUnknown)
			return
		}

		tokenString, expTime, err := EncodeToken(user.ID, user.JWTSecret, issuer, ttl)
		if err != nil {
			r.Logf("encoding user=%s issuer=%", user.ID, issuer)
			r.DenyAccessForUser(user.ID, AuthenticationError, JWTEncodingFailure)
			return
		}

		token := JWTToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}
		c.JSON(http.StatusOK, token)

		r.RecordEvent(user.ID, "login_successful")
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
func GetUser(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// user granted access to his/her own account
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		id := c.Param("id")
		serr := r.CheckUserIDMatchUserFromContext(id)
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, r.WrapAsJSONAPIErrors(serr))
			return
		}

		user, dberr := userService.Get(id)
		if dberr != nil {
			r.Logf("user.get-profile.failed err=%s", dberr)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}

		r.SetContentTypeJSON()
		out, err := MarshalUser(c.Request.URL.RequestURI(), UserJSONSchema, user)
		if err != nil {
			r.Logf("user.get-profile.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
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
func DeleteUser(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		id := c.Param("id")
		serr := r.CheckUserIDMatchUserFromContext(id)
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, r.WrapAsJSONAPIErrors(serr))
			return
		}

		dberr := userService.Delete(id)
		if dberr != nil {
			r.Logf("user.delete.failed id=%s error=%s", id, dberr)
			r.AbortWithStatusUnauthorizedError(DatabaseError, DeletingFailed)
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
func UpdateUser(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		id := c.Param("id")
		serr := r.CheckUserIDMatchUserFromContext(id)
		if serr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, r.WrapAsJSONAPIErrors(serr))
			return
		}

		payload, err := r.GetBody()
		if err != nil {
			r.AbortWithStatusInternalServerError(ServerError, RetrievingPayloadError)
			return
		}

		user, err := UnmarshalUser(payload, UserJSONSchema)
		if err != nil {
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
			return
		}

		dberr := userService.UpdateProfile(user)
		if dberr != nil {
			r.Logf("signup.profile-update.failed err=%s", dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, GeneralError)
		}

		c.Status(http.StatusNoContent)
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
