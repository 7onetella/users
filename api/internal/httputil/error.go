package httputil

import (
	"log"
)

type DBOpError struct {
	Query string
	Err   error
}

func (e *DBOpError) Unwrap() error {
	return e.Err
}

func (e *DBOpError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

func (e *DBOpError) Log(tx string) {
	log.Printf("%s sql.excute.errored: %#v, sql: %s", tx, e.Err, e.Query)
}

func LogErr(txid string, message string, opserr interface{}) {
	log.Printf("%s %s: %#v", txid, message, opserr)
}

func LogDBErr(txid string, sql, message string, opserr interface{}) {
	log.Printf("%s %s: %#v, sql: %s", txid, message, opserr, sql)
}
