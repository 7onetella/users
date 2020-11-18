package httputil

import (
	"github.com/7onetella/users/api/internal/model"
	"log"
)

type DBOpError struct {
	Query string
	Err   error
}

func (e *DBOpError) Unwrap() error {
	return e.Err
}

func (e *DBOpError) Log(tx string) {
	log.Printf("%s sql.excute.errored: %#v, sql: %s", tx, e.Err, e.Query)
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

func (rh RequestHandler) HandleDBError(dberr *DBOpError) bool {
	if dberr != nil {
		c := rh.Context
		LogDBErr(rh.TX(), dberr.Query, "db error", dberr.Err)
		c.AbortWithStatus(500)
		return true
	}
	return false
}

func (rh RequestHandler) WrapAsJSONAPIErrors(err *model.Error) model.JSONAPIErrors {
	if err == nil {
		return model.JSONAPIErrors{}
	}

	out := model.JSONAPIErrors{
		Errors: []model.JSONAPIError{
			{
				Meta: err,
			},
		},
	}
	return out
}

func LogErr(txid string, message string, opserr interface{}) {
	log.Printf("%s %s: %#v", txid, message, opserr)
}

func LogDBErr(txid string, sql, message string, opserr interface{}) {
	log.Printf("%s %s: %#v, sql: %s", txid, message, opserr, sql)
}
