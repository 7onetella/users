package httputil

import (
	"log"
)

func LogErr(txid string, message string, opserr interface{}) {
	log.Printf("%s %s: %#v", txid, message, opserr)
}

func LogDBErr(txid string, sql, message string, opserr interface{}) {
	log.Printf("%s %s: %#v, sql: %s", txid, message, opserr, sql)
}
