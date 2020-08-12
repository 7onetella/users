package main

import (
	"context"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

var db *sqlx.DB

func init() {
	newdb, err := sqlx.Connect("postgres", "host=tmt-vm11.7onetella.net user=dev password=dev114 dbname=devdb sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	db = newdb

	db.MustExec(DBSchema)
}

func main() {

	r := gin.Default()

	r.Use(PreflightOptions())

	r.Use(TransactionID())

	userService := UserService{DB: db}

	jwt := JWTAuth{
		ClaimKey: "user_id",
		TTL:      120,
	}

	usersG := r.Group("/users")
	usersG.Use(jwt.Validator(userService))
	{
		usersG.GET("/:id", GetUser(userService))
		usersG.DELETE("/:id", DeleteUser(userService))
		usersG.PATCH("/:id", UpdateUser(userService))
	}
	r.POST("/users", Signup(userService))

	mfa := r.Group("/totp")
	mfa.Use(jwt.Validator(userService))
	{
		mfa.GET("/qr-code-raw", NewMFA(userService))
		mfa.GET("/qr-code-json", NewMFABase64(userService))
		mfa.POST("/confirm", ConfirmToken(userService))
	}

	auth := r.Group("/jwt_auth")
	{
		auth.POST("/signin", jwt.Signin(userService))
		auth.POST("/refresh", jwt.RefreshToken(userService))
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
