package main

import (
	"context"
	_ "github.com/7onetella/users/api/docs"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/handlers"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"log"
	"net/http"
)

// @title Users API
// @version 1.0
// @description API for totp and webauthn auth with user self mgt api

// @contact.name 7onetella
// @contact.url https://github.com/7onetella/users

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /users
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @x-extension-openapi {"example": "value on a json format"}

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
		TTL:      3600,
	}

	web, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "7onetella", // display name for your site
		RPID:          "localhost", // generally the domain name for your site
		RPOrigin:      "http://localhost:4200",
	})
	if err != nil {
		log.Fatal(err)
	}

	h := http.StripPrefix("/ui/", http.FileServer(assetFS()))
	r.GET("/ui/*path", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})

	users := r.Group("/users")
	users.Use(jwt.TokenValidator(userService))
	{
		users.GET("/:id", GetUser(userService))
		users.DELETE("/:id", DeleteUser(userService))
		users.PATCH("/:id", UpdateUser(userService))
	}
	r.POST("/users", Signup(userService))

	totp := r.Group("/totp")
	totp.Use(jwt.TokenValidator(userService))
	{
		totp.GET("/qr-code-raw", NewTOTPRaw(userService))
		totp.GET("/qr-code-json", NewTOTPJson(userService))
		totp.POST("/confirm", ConfirmToken(userService))
	}

	auth := r.Group("/jwt_auth")
	{
		auth.POST("/signin", Signin(userService, jwt.ClaimKey, jwt.TTL))
		auth.POST("/refresh", jwt.RefreshToken(userService))
	}

	webauthn := r.Group("/webauthn")
	webauthn.Use(jwt.TokenValidator(userService))
	{
		webauthn.POST("/register/begin", BeginRegistration(userService, web))
		webauthn.POST("/register/finish", FinishRegistration(userService, web))
	}
	r.POST("/webauthn/login/begin", BeginLogin(userService, web))
	r.POST("/webauthn/login/finish", FinishLogin(userService, web))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
