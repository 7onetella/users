package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mfcochauxlaberge/jsonapi"
	"log"
)

var db *sqlx.DB
var userSchema *jsonapi.Schema

func init() {
	newdb, err := sqlx.Connect("postgres", "host=tmt-vm11.7onetella.net user=dev password=dev114 dbname=devdb sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	db = newdb

	s, errs := SchemaCheck(&User{})
	if errs != nil && len(errs) > 0 {
		log.Fatalln(errs)
	}
	userSchema = s

	db.MustExec(dbschema)
}

func main() {

	r := gin.Default()

	r.Use(PreflightOptions())

	r.Use(TransactionID())

	//secret := "abcdefg"

	userService := UserService{db}

	jwt := JWTAuth{
		//secret:   secret,
		claimKey: "user_id",
		ttl:      120,
	}

	usersG := r.Group("/users")
	usersG.Use(jwt.Validator(userService))
	{
		//usersG.GET("", ListUsers(userService))
		usersG.GET("/:id", GetUser(userService))
		usersG.DELETE("/:id", DeleteUser(userService))
		usersG.PATCH("/:id", UpdateUser(userService))
	}
	r.POST("/users", Signup(userService))

	mfa := r.Group("/mfa")
	mfa.Use(jwt.Validator(userService))
	{
		mfa.GET("/new", NewMFA(userService))

		mfa.GET("/new/png_base64", NewMFABase64(userService))

		mfa.POST("/confirm", ConfirmToken(userService))

		mfa.GET("/otp", GenerateOTP())
	}

	auth := r.Group("/jwt_auth")
	{
		auth.POST("/signin", jwt.Signin(userService))

		auth.POST("/refresh", jwt.RefreshToken())
	}

	r.Run()
}

func TransactionID() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := uuid.New().String()
		req := c.Request
		url := req.URL

		ctx := context.WithValue(req.Context(), "tid", tid)
		c.Request = c.Request.Clone(ctx)
		c.Request.URL = url
		c.Next()
	}
}
