package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mfcochauxlaberge/jsonapi"
	"io/ioutil"
	"log"
)

// txid, logging, cors header, payload
type RequestHanlder struct {
	Context *gin.Context
	Errors  []error
}

func NewRequestHandler(c *gin.Context) RequestHanlder {
	return RequestHanlder{
		c,
		[]error{},
	}
}

func (rh RequestHanlder) WriteCORSHeader() {
	w := rh.Context.Writer
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func (rh RequestHanlder) SetContentTypeJSON() {
	c := rh.Context
	c.Header("Content-Type", "application/json")
}

func (rh RequestHanlder) GetBody() ([]byte, error) {
	c := rh.Context
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	log.Println("payload = %s", string(payload))
	return payload, err
}

func (rh RequestHanlder) SchemaCheck(v interface{}) (*jsonapi.Schema, []error) {
	schema := &jsonapi.Schema{}
	schema.AddType(jsonapi.MustBuildType(v))
	errors := schema.Check()
	return schema, errors
}

func SchemaCheck(v interface{}) (*jsonapi.Schema, []error) {
	schema := &jsonapi.Schema{}
	schema.AddType(jsonapi.MustBuildType(v))
	errors := schema.Check()
	return schema, errors
}

func (rh RequestHanlder) HandleError(errs ...error) bool {
	c := rh.Context

	if errs == nil || len(errs) == 0 {
		return false
	}

	errFound := false
	for i, _ := range errs {
		err := errs[i]
		if err != nil {
			LogErr(rh.TransactionIDFromContext(), "errs[i]", err)
			errFound = true
		}
	}
	if errFound {
		c.AbortWithStatus(500)
		return true
	}
	return false
}

func (rh RequestHanlder) HandleDBError(dberr *DBOpError) bool {
	if dberr != nil {
		c := rh.Context
		LogErr(rh.TransactionIDFromContext(), "db error", dberr.Err)
		c.AbortWithStatus(500)
		return true
	}
	return false
}

func (rh RequestHanlder) HandleSecurityError(serr *SecurityError) bool {
	if serr != nil {
		c := rh.Context
		LogErr(rh.TransactionIDFromContext(), "security error", serr)
		c.AbortWithStatus(401)
		return true
	}
	return false
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
