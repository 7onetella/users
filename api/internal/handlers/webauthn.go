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

func BeginRegistration(service UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		user, err := rh.UserFromContext()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(AuthenticationError, UserUnknown))
			return
		}

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginRegistration(user)
		if err != nil {
			LogErr(rh.TX(), "error registering user", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(AuthenticationError, WebauthnRegistrationFailure))
			return
		}

		marshaledData, err := json.Marshal(sessionData)
		if err != nil {
			rh.AbortWithStatusInternalServerError(JSONAPISpecError, Marshalling, err)
			return
		}

		user.WebAuthnSessionData = string(marshaledData)
		dberr := service.SaveWebauthnSession(user)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := json.Marshal(options)
		if err != nil {
			LogErr(rh.TX(), "error marshalling PublicKeyCredentialCreationOptions", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Marshalling))
			return
		}
		c.String(http.StatusOK, string(out))
	}
}

func FinishRegistration(service UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		user, err := rh.UserFromContext()
		if rh.HandleError(err) {
			return
		}

		marshaledData := []byte(user.WebAuthnSessionData)
		sessionData := webauthn.SessionData{}
		err = json.Unmarshal(marshaledData, &sessionData)
		if rh.HandleError(err) {
			return
		}

		credential, err := web.FinishRegistration(user, sessionData, c.Request)
		if rh.HandleError(err) {
			return
		}

		marshaledCredential, err := json.Marshal(credential)
		if rh.HandleError(err) {
			return
		}

		service.SaveUserCredential(NewUserCredential(user.ID, string(marshaledCredential)))

		user.WebAuthnEnabled = true
		// clean up the session data after successful registration
		user.WebAuthnSessionData = ""
		service.UpdateWebAuthn(user)

		c.JSON(200, gin.H{"result": true})
	}
}

func BeginLogin(service UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		signinSessionToken := c.GetHeader(SigninSessionTokenHeader)
		eventID, _ := Base64Decode(signinSessionToken)
		log.Printf("event id = %s", eventID)
		user, dberr := service.FindUserByAuthEventID(eventID)
		if rh.HandleDBError(dberr) {
			return
		}
		log.Printf("user = %#v", user)

		userCredentials, dberr := service.FindUserCredentialByUserID(user.ID)
		if rh.HandleDBError(dberr) {
			return
		}
		var credentials []webauthn.Credential
		for _, cred := range userCredentials {
			credential := webauthn.Credential{}
			err := json.Unmarshal([]byte(cred.Credential), &credential)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Unmarshalling))
				return
			}
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginLogin(user)
		if rh.HandleError(err) {
			return
		}

		marshaledData, err := json.Marshal(sessionData)
		if rh.HandleError(err) {
			return
		}
		user.WebAuthnSessionData = string(marshaledData)
		dberr = service.SaveWebauthnSession(user)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := json.Marshal(&options)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, string(out))
	}
}

func FinishLogin(service UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		signInSessionToken := c.GetHeader(SigninSessionTokenHeader)
		eventID, _ := Base64Decode(signInSessionToken)
		user, dberr := service.FindUserByAuthEventID(eventID)
		if rh.HandleDBError(dberr) {
			return
		}

		marshaledData := []byte(user.WebAuthnSessionData)
		sessionData := webauthn.SessionData{}
		err := json.Unmarshal(marshaledData, &sessionData)
		if rh.HandleError(err) {
			return
		}

		userCredentials, dberr := service.FindUserCredentialByUserID(user.ID)
		if rh.HandleDBError(dberr) {
			return
		}
		var credentials []webauthn.Credential
		for _, cred := range userCredentials {
			credential := webauthn.Credential{}
			err = json.Unmarshal([]byte(cred.Credential), &credential)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Unmarshalling))
				return
			}
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		_, err = web.FinishLogin(user, sessionData, c.Request)
		if rh.HandleError(err) {
			return
		}

		secAuthEvent := NewAuthEvent(user.ID, "sec_auth_success", "", "", "")
		service.RecordAuthEvent(secAuthEvent)

		c.JSON(http.StatusOK, gin.H{
			"result":                 true,
			"signin_session_token":   signInSessionToken,
			"webauthn_session_token": Base64Encode(secAuthEvent.ID),
		})
	}
}
