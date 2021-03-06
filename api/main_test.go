package main

import (
	"github.com/7onetella/users/api/internal/dbutil"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"testing"
)

var testDB *sqlx.DB

func setup() {
	db, err := sqlx.Connect("postgres", "host=tmt-vm11.7onetella.net user=dev password=dev114 dbname=devdb sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	testDB = db

	db.MustExec(dbutil.DBSchema)
}

func teardown() {

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
