package httputil

import (
	"errors"
	. "github.com/7onetella/users/api/internal/model"
	"io/ioutil"
	"log"
	"net/http"
)

func (rh RequestHandler) NewAuthEvent(userID, event string) AuthEvent {
	c := rh.Context

	return NewAuthEvent(userID, event, c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
}

func (rh RequestHandler) GetBody() ([]byte, error) {
	c := rh.Context
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	return payload, err
}

func (rh RequestHandler) UserFromContext() (User, error) {
	c := rh.Context
	ctx := c.Request.Context()
	user, ok := ctx.Value("user").(User)
	if !ok {
		return User{}, errors.New("error getting user from context")
	}
	return user, nil
}

// TX returns request transaction id
func (rh RequestHandler) TX() string {
	c := rh.Context
	return c.Request.Context().Value("tid").(string)
}

func (rh RequestHandler) CheckUserIDMatchUserFromContext(id string) *Error {
	ctxUser, err := rh.UserFromContext()
	if err != nil {
		return New(SecurityError, Unknown)
	}

	if id != ctxUser.ID {
		return New(SecurityError, ContextUserDoesNotMatchGivenUserID)
	}

	return nil
}

func (rh RequestHandler) WrapAsJSONAPIErrors(err *Error) JSONAPIErrors {
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

func (rh RequestHandler) HandleDBError(dberr *DBOpError) bool {
	if dberr != nil {
		c := rh.Context
		e := Wrap(DatabaseError, GeneralError, dberr)
		rh.LogError(e)
		c.AbortWithStatus(http.StatusInternalServerError)
		return true
	}
	return false
}

func (rh RequestHandler) HandleError(errs ...error) bool {
	c := rh.Context

	if errs == nil || len(errs) == 0 {
		return false
	}

	errFound := false
	for i := range errs {
		err := errs[i]
		if err != nil {
			LogErr(rh.TX(), "errs[i]", err)
			errFound = true
		}
	}
	if errFound {
		c.AbortWithStatus(500)
		return true
	}
	return false
}

func (rh RequestHandler) Log(message string) {
	log.Printf("%s %s", rh.TX(), message)
}

func (rh RequestHandler) Logf(format string, v ...interface{}) {
	var message []interface{}
	message = append(message, rh.TX())
	message = append(message, v...)
	log.Printf("%s "+format, message...)
}

func (rh RequestHandler) LogError(e *Error) {
	// f7fc7487-9b92-4f27-8d4f-83ca96bbb6b9 {"code":3200,"reason":"","message":"Unable to marshall json"}
	log.Printf("%s error=%s", rh.TX(), e)
}

func (rh RequestHandler) AbortWithStatusInternalServerError(category Category, reason Reason) {
	e := New(category, reason)
	rh.Context.AbortWithStatusJSON(http.StatusInternalServerError, e)
}
