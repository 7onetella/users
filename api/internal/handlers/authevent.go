package handlers

import (
	"net/http"
	"time"

	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/crypto"
	"github.com/gin-gonic/gin"
)

type AuthEventHandler struct {
	RequestHandler RequestHandler
	UserService    UserService
}

func (ah AuthEventHandler) Context() *gin.Context {
	return ah.RequestHandler.Context
}

func (ah AuthEventHandler) DenyAccessForAnonymous(category Category, reason Reason) {
	ah.DenyAccessForUser("", category, reason)
}

func (ah AuthEventHandler) DenyAccessForUser(userID string, category Category, reason Reason) {
	c := ah.Context()
	e := New(category, reason)
	c.AbortWithStatusJSON(http.StatusUnauthorized, e)
	// record potential attacker's ip address and browser agent
	ah.RecordEvent(userID, e.Message)
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
	rh := ah.RequestHandler
	event := rh.NewAuthEvent(userID, eventName)
	dberr := ah.UserService.RecordAuthEvent(event)
	if dberr != nil {
		rh.LogError(Wrap(DatabaseError, PersistingFailed, dberr))
	}
}

func (ah AuthEventHandler) AccessDeniedMissingData(userID string, category Category, reason Reason) {
	rh := ah.RequestHandler
	c := rh.Context
	e := New(category, reason)
	userService := ah.UserService

	//e := New(category, reason)
	event := rh.NewAuthEvent(userID, e.Message)
	userService.RecordAuthEvent(event)

	mde := MissingDataError{
		Error:              *e,
		SigninSessionToken: crypto.Base64Encode(event.ID),
	}

	c.AbortWithStatusJSON(422, mde)
}

func (ah AuthEventHandler) FinishSecondAuth(userID, eventName, message string) {
	rh := ah.RequestHandler
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

func (ah AuthEventHandler) IsWebAuthnSessionTokenValidForUer(userID, webAuthnSessionToken string) bool {
	userService := ah.UserService
	eventID, _ := crypto.Base64Decode(webAuthnSessionToken)
	user, dberr := userService.FindUserByAuthEventID(eventID)
	if dberr != nil {
		return false
	}

	return userID == user.ID
}
