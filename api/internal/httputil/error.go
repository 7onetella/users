package httputil

import (
	"encoding/json"
	"log"
)

type DBOpError struct {
	Query string `json:"query"`
	Err   error  `json:"error"`
}

func (e *DBOpError) Unwrap() error {
	return e.Err
}

// The error interface implementation, which formats to a JSON object string.
func (e *DBOpError) Error() string {
	marshaled, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(marshaled)
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
