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
	"os"
	"path/filepath"
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
var stage string
var port string
var _RPID string
var _RPOrigin string
var connStr string

func init() {
	stage = GetEnvWithDefault("STAGE", "localhost")

	port = ":" + GetEnvWithDefault("HTTP_PORT", "8080")

	_RPID = GetEnvWithDefault("RPID", "localhost")

	_RPOrigin = GetEnvWithDefault("RPORIGIN", "http://localhost:4200")

	connStr = GetEnvWithDefault("DB_CONNSTR", "host=tmt-vm11.7onetella.net user=dev password=dev114 dbname=devdb sslmode=disable")

	newdb, err := sqlx.Connect("postgres", connStr)
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
		RPDisplayName: _RPID, // display name for your site
		RPID:          _RPID, // generally the domain name for your site
		RPOrigin:      _RPOrigin,
	})
	if err != nil {
		log.Fatal(err)
	}

	// ----- EmberJS SPA resource ---------------------------------
	h := http.StripPrefix("/accounts/", http.FileServer(assetFS()))
	r.GET("/accounts/*path", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	})

	// ----- User Profile -----------------------------------------
	users := r.Group("/users")
	users.Use(jwt.TokenValidator(userService))
	{
		users.GET("/:id", GetUser(userService))
		users.DELETE("/:id", DeleteUser(userService))
		users.PATCH("/:id", UpdateUser(userService))
	}
	r.POST("/users", Signup(userService))

	// ----- User Authentication ----------------------------------
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

	// ---- OAuth2 ---------------------------------
	oauth2author := r.Group("/oauth2")
	oauth2author.Use(jwt.TokenValidator(userService))
	{
		oauth2author.POST("/authorize", OAuth2Authorize(userService))
	}
	r.POST("/oauth2/access_token", OAuth2AccessToken(userService))

	// ---- Client --------------------------------
	r.GET("/oauth2/clients/:id", GetClientName(userService))

	// ---- Swagger Documentation ---------------------------------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("rpid = %s, rporigin = %s \n", _RPID, _RPOrigin)

	switch stage {
	case "localhost":
		r.Run(port)
	case "live":
		certFile, keyFile := GetCertAndKey()
		r.RunTLS(port, certFile, keyFile)
	default:
	}

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

// GetCertAndKey returns cert and key locations
func GetCertAndKey() (string, string) {
	// this will resolve to refresh or api
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("workfolder:", dir)

	return dir + "/" + stage + "-crt.pem", dir + "/" + stage + "-key.pem"
}

// GetEnvWithDefault attemps to retrieve from env. default calculated based on stage if env value empty.
func GetEnvWithDefault(env, defaultV string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultV
	}
	return v
}
