package handlers

import (
	"encoding/json"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	. "github.com/7onetella/users/api/pkg/crypto"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

const SigninSessionTokenHeader = "SignSessionToken"

func BeginRegistration(userService UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		user, err := r.UserFromContext()
		if err != nil {
			r.AbortWithStatusInternalServerError(AuthenticationError, UserUnknown)
			return
		}

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginRegistration(user)
		if err != nil {
			r.Logf("signup.begin-registration.failed err=%s", err)
			r.AbortWithStatusInternalServerError(AuthenticationError, WebauthnRegistrationFailure)
			return
		}

		marshaledData, err := json.Marshal(sessionData)
		if err != nil {
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}

		user.WebAuthnSessionData = string(marshaledData)
		dberr := userService.SaveWebauthnSession(user)
		if dberr != nil {
			r.Logf("signup.begin-registration.failed err=%s", dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, GeneralError)
		}

		r.SetContentTypeJSON()
		response, err := json.Marshal(options)
		if err != nil {
			r.Logf("webauthn.begin-registration.failed err=%s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Marshalling))
			return
		}
		c.String(http.StatusOK, string(response))
	}
}

func FinishRegistration(userService UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		user, err := r.UserFromContext()
		if err != nil {
			r.Logf("webauthn.finish-registration.failed err=%s", err)
			r.AbortWithStatusInternalServerError(AuthenticationError, UserUnknown)
			return
		}

		marshaledData := []byte(user.WebAuthnSessionData)
		sessionData := webauthn.SessionData{}
		err = json.Unmarshal(marshaledData, &sessionData)
		if err != nil {
			r.Logf("webauthn.finish-registration.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
			return
		}

		credential, err := web.FinishRegistration(user, sessionData, c.Request)
		if err != nil {
			r.Logf("webauthn.finish-registration.failed err=%s", err)
			r.AbortWithStatusInternalServerError(WebauthnError, RegistrationError)
			return
		}

		marshaledCredential, err := json.Marshal(credential)
		if err != nil {
			r.Logf("webauthn.finish-registration.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}

		userService.SaveUserCredential(NewUserCredential(user.ID, string(marshaledCredential)))

		user.WebAuthnEnabled = true
		// clean up the session data after successful registration
		user.WebAuthnSessionData = ""
		userService.UpdateWebAuthn(user)

		c.JSON(200, gin.H{"result": true})
	}
}

func BeginLogin(userService UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		signinSessionToken := c.GetHeader(SigninSessionTokenHeader)
		eventID, _ := Base64Decode(signinSessionToken)
		log.Printf("event id = %s", eventID)
		user, dberr := userService.FindUserByAuthEventID(eventID)
		if dberr != nil {
			r.Logf("webauthn.finding-user.failed event_id=%s error=%s", eventID, dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, QueryingFailed)
			return
		}

		userCredentials, dberr := userService.FindUserCredentialByUserID(user.ID)
		if dberr != nil {
			r.Logf("webauthn.finding-user-credential.failed user_id=%s error=%s", user.ID, dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, QueryingFailed)
			return
		}
		var credentials []webauthn.Credential
		for _, cred := range userCredentials {
			credential := webauthn.Credential{}
			err := json.Unmarshal([]byte(cred.Credential), &credential)
			if err != nil {
				r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
				return
			}
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginLogin(user)
		if err != nil {
			r.Logf("webauthn.begin-login.failed err=%s", err)
			r.AbortWithStatusInternalServerError(WebauthnError, LoginFailed)
			return
		}

		marshaledData, err := json.Marshal(sessionData)
		if err != nil {
			r.Logf("webauthn.begin-login.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}
		user.WebAuthnSessionData = string(marshaledData)
		dberr = userService.SaveWebauthnSession(user)
		if dberr != nil {
			r.AbortWithStatusInternalServerError(DatabaseError, PersistingFailed)
			return
		}

		r.SetContentTypeJSON()
		response, err := json.Marshal(&options)
		if err != nil {
			r.Logf("webauthn.begin-login.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling)
			return
		}
		c.String(http.StatusOK, string(response))
	}
}

func FinishLogin(userService UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		signInSessionToken := c.GetHeader(SigninSessionTokenHeader)
		eventID, _ := Base64Decode(signInSessionToken)
		user, dberr := userService.FindUserByAuthEventID(eventID)
		if dberr != nil {
			r.Logf("webauthn.finding-auth-event.failed event_id=%s error=%s", eventID, dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, QueryingFailed)
			return
		}

		marshaledData := []byte(user.WebAuthnSessionData)
		sessionData := webauthn.SessionData{}
		err := json.Unmarshal(marshaledData, &sessionData)
		if err != nil {
			r.Logf("webauthn.finish-login.failed err=%s", err)
			r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
			return
		}

		userCredentials, dberr := userService.FindUserCredentialByUserID(user.ID)
		if dberr != nil {
			r.Logf("webauthn.finding-credentials-for-user.failed usert_id=%s error=%s", user.ID, dberr)
			r.AbortWithStatusInternalServerError(DatabaseError, QueryingFailed)
			return
		}
		var credentials []webauthn.Credential
		for _, cred := range userCredentials {
			credential := webauthn.Credential{}
			err = json.Unmarshal([]byte(cred.Credential), &credential)
			if err != nil {
				r.AbortWithStatusInternalServerError(JSONAPISpecError, Unmarshalling)
				return
			}
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		_, err = web.FinishLogin(user, sessionData, c.Request)
		if err != nil {
			r.Logf("webauthn.finish-login.failed err=%s", err)
			r.AbortWithStatusInternalServerError(WebauthnError, FinishLoginError)
			return
		}

		secAuthEvent := NewAuthEvent(user.ID, "sec_auth_success", "", "", "")
		userService.RecordAuthEvent(secAuthEvent)

		c.JSON(http.StatusOK, gin.H{
			"result":                 true,
			"signin_session_token":   signInSessionToken,
			"webauthn_session_token": Base64Encode(secAuthEvent.ID),
		})
	}
}
