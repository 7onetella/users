package httputil

import (
	"errors"
	. "github.com/7onetella/users/api/internal/model"
	"io/ioutil"
	"log"
)

func (rh RequestHanlder) GetBody() ([]byte, error) {
	c := rh.Context
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	log.Println("payload = %s", string(payload))
	return payload, err
}

func (rh RequestHanlder) UserFromContext() (User, error) {
	c := rh.Context
	ctx := c.Request.Context()
	user, ok := ctx.Value("user").(User)
	if !ok {
		return User{}, errors.New("error getting user from context")
	}
	return user, nil
}

func (rh RequestHanlder) TransactionIDFromContext() string {
	c := rh.Context
	return c.Request.Context().Value("tid").(string)
}

func (rh RequestHanlder) GetPayload(v interface{}) ([]byte, []error) {

	payload, err := rh.GetBody()
	if err != nil {
		return nil, []error{err}
	}

	return payload, nil
}

func (rh RequestHanlder) CheckUserIDMatchUserFromContext(id string) *SecurityError {
	ctxUser, err := rh.UserFromContext()
	if err != nil {
		return &SecurityError{
			Event: "get user from context",
			Err:   err,
		}
	}

	if id != ctxUser.ID {
		return &SecurityError{
			Event: "user id check",
			Err:   errors.New("id from url parameter does not match user id from context"),
		}
	}

	return nil
}
