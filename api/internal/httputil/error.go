package httputil

import "log"

type DBOpError struct {
	Query string
	Err   error
}

func (e *DBOpError) Unwrap() error {
	return e.Err
}

type SecurityError struct {
	Event string
	Err   error
}

func (e *SecurityError) Unwrap() error {
	return e.Err
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

func LogErr(txid string, message string, opserr interface{}) {
	log.Printf("%s %s: %#v", txid, message, opserr)
}
