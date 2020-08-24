package handlers

import (
	"encoding/json"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/7onetella/users/api/internal/model"
	. "github.com/7onetella/users/api/pkg/crypto"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func BeginRegistration(service UserService, web *webauthn.WebAuthn) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		user, err := rh.UserFromContext()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginRegistration(
			user,
		)
		if rh.HandleError(err) {
			return
		}

		marshaledData, err := json.Marshal(sessionData)
		if rh.HandleError(err) {
			return
		}
		user.WebAuthnSessionData = string(marshaledData)
		dberr := service.SaveWebauthnSession(user)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := json.Marshal(options)
		if rh.HandleError(err) {
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

		service.SaveUserCredential(model.NewUserCredential(user.ID, string(marshaledCredential)))

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

		authToken := c.GetHeader("AuthToken")
		eventID, _ := Base64Decode(authToken)
		log.Printf("event id = %s", eventID)
		user, dberr := service.FindUserByAuthEventID(eventID)
		if rh.HandleDBError(dberr) {
			return
		}
		log.Printf("user = %#v", user)

		userCreds, dberr := service.FindUserCredentialByUserID(user.ID)
		if rh.HandleDBError(dberr) {
			return
		}
		credentials := []webauthn.Credential{}
		for _, cred := range userCreds {
			credential := webauthn.Credential{}
			json.Unmarshal([]byte(cred.Credential), &credential)
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		// generate PublicKeyCredentialCreationOptions, session data
		options, sessionData, err := web.BeginLogin(
			user,
		)
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

		authToken := c.GetHeader("AuthToken")
		eventID, _ := Base64Decode(authToken)
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

		userCreds, dberr := service.FindUserCredentialByUserID(user.ID)
		if rh.HandleDBError(dberr) {
			return
		}
		credentials := []webauthn.Credential{}
		for _, cred := range userCreds {
			credential := webauthn.Credential{}
			json.Unmarshal([]byte(cred.Credential), &credential)
			credentials = append(credentials, credential)
		}
		user.Credentials = credentials

		// TODO: perform additional check
		_, err = web.FinishLogin(user, sessionData, c.Request)
		if rh.HandleError(err) {
			return
		}

		secAuthEvent := model.NewAuthEvent(user.ID, "sec_auth_success", "", "", "")
		service.RecordAuthEvent(secAuthEvent)

		c.JSON(200, gin.H{
			"result":         true,
			"auth_token":     authToken,
			"sec_auth_token": Base64Encode(secAuthEvent.ID),
		})
	}
}
