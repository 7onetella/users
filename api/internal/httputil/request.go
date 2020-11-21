package httputil

import (
	"errors"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/crypto"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// txid, logging, cors header, payload
type RequestHandler struct {
	Context     *gin.Context
	UserService UserService
	Errors      []error
}

func NewRequestHandler(c *gin.Context, userService UserService) RequestHandler {
	return RequestHandler{
		c,
		userService,
		[]error{},
	}
}

func (r RequestHandler) NewAuthEvent(userID, event string) AuthEvent {
	c := r.Context

	return NewAuthEvent(userID, event, c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
}

func (r RequestHandler) GetBody() ([]byte, error) {
	c := r.Context
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	return payload, err
}

func (r RequestHandler) UserFromContext() (User, error) {
	c := r.Context
	ctx := c.Request.Context()
	user, ok := ctx.Value("user").(User)
	if !ok {
		return User{}, errors.New("error getting user from context")
	}
	return user, nil
}

// TX returns request transaction id
func (r RequestHandler) TX() string {
	c := r.Context
	return c.Request.Context().Value("tid").(string)
}

func (r RequestHandler) CheckUserIDMatchUserFromContext(id string) *Error {
	ctxUser, err := r.UserFromContext()
	if err != nil {
		return New(SecurityError, Unknown)
	}

	if id != ctxUser.ID {
		return New(SecurityError, ContextUserDoesNotMatchGivenUserID)
	}

	return nil
}

func (r RequestHandler) WrapAsJSONAPIErrors(err *Error) JSONAPIErrors {
	if err == nil {
		return JSONAPIErrors{}
	}

	out := JSONAPIErrors{
		Errors: []JSONAPIError{
			{
				Meta: err,
			},
		},
	}
	return out
}

func (r RequestHandler) HandleDBError(dberr *DBOpError) bool {
	if dberr != nil {
		c := r.Context
		e := Wrap(DatabaseError, GeneralError, dberr)
		r.LogError(e)
		c.AbortWithStatus(http.StatusInternalServerError)
		return true
	}
	return false
}

func (r RequestHandler) HandleError(errs ...error) bool {
	c := r.Context

	if errs == nil || len(errs) == 0 {
		return false
	}

	errFound := false
	for i := range errs {
		err := errs[i]
		if err != nil {
			LogErr(r.TX(), "errs[i]", err)
			errFound = true
		}
	}
	if errFound {
		c.AbortWithStatus(500)
		return true
	}
	return false
}

func (r RequestHandler) Log(message string) {
	log.Printf("%s %s", r.TX(), message)
}

func (r RequestHandler) Logf(format string, v ...interface{}) {
	var message []interface{}
	message = append(message, r.TX())
	message = append(message, v...)
	log.Printf("%s "+format, message...)
}

func (r RequestHandler) LogError(e *Error) {
	// f7fc7487-9b92-4f27-8d4f-83ca96bbb6b9 {"code":3200,"reason":"","message":"Unable to marshall json"}
	log.Printf("%s error=%s", r.TX(), e)
}

func (r RequestHandler) AbortWithStatusInternalServerError(category Category, reason Reason) {
	e := New(category, reason)
	r.Context.AbortWithStatusJSON(http.StatusInternalServerError, e)
}

func (r RequestHandler) DenyAccessForAnonymous(category Category, reason Reason) {
	r.DenyAccessForUser("", category, reason)
}

func (r RequestHandler) DenyAccessForUser(userID string, category Category, reason Reason) {
	c := r.Context
	e := New(category, reason)
	c.AbortWithStatusJSON(http.StatusUnauthorized, e)
	// record potential attacker's ip address and browser agent
	r.RecordEvent(userID, e.Message)
}

func (r RequestHandler) ExtractEventID(s string) (string, error) {
	decoded, err := crypto.Base64Decode(s)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func (r RequestHandler) IsSigninSessionStillValid(timestamp int64, limit time.Duration) bool {
	return (time.Now().Unix() - timestamp) > int64(limit.Seconds())
}

func (r RequestHandler) RecordEvent(userID, eventName string) {
	event := r.NewAuthEvent(userID, eventName)
	dberr := r.UserService.RecordAuthEvent(event)
	if dberr != nil {
		r.LogError(Wrap(DatabaseError, PersistingFailed, dberr))
	}
}

func (r RequestHandler) AccessDeniedMissingData(userID string, category Category, reason Reason) {
	c := r.Context
	e := New(category, reason)
	userService := r.UserService

	//e := New(category, reason)
	event := r.NewAuthEvent(userID, e.Message)
	userService.RecordAuthEvent(event)

	mde := MissingDataError{
		Error:              *e,
		SigninSessionToken: crypto.Base64Encode(event.ID),
	}

	c.AbortWithStatusJSON(422, mde)
}

func (r RequestHandler) FinishSecondAuth(userID, eventName, message string) {
	c := r.Context
	userService := r.UserService

	event := r.NewAuthEvent(userID, eventName)
	userService.RecordAuthEvent(event)

	secEvent := r.NewAuthEvent(userID, "sec_auth_generated")
	userService.RecordAuthEvent(secEvent)

	c.AbortWithStatusJSON(401, gin.H{
		"reason":                 event.Event,
		"message":                message,
		"signin_session_token":   crypto.Base64Encode(event.ID),
		"webauthn_session_token": crypto.Base64Encode(secEvent.ID),
	})
}

func (r RequestHandler) IsWebAuthnSessionTokenValidForUer(userID, webAuthnSessionToken string) bool {
	userService := r.UserService
	eventID, _ := crypto.Base64Decode(webAuthnSessionToken)
	user, dberr := userService.FindUserByAuthEventID(eventID)
	if dberr != nil {
		return false
	}

	return userID == user.ID
}

func (r RequestHandler) WriteCORSHeader() {
	w := r.Context.Writer
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func (r RequestHandler) SetContentTypeJSON() {
	c := r.Context
	c.Header("Content-Type", "application/json")
}
