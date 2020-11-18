// Package classification Users API.
//
// API for secure authentication and oauth2 authorization
//
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: https
//     Host: accounts.7onetella.net
//     BasePath: /
//     Version: 0.1.0
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: Seven Tella<7onetella@gmail.com>
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - bearer_token:
//
//     SecurityDefinitions:
//     bearer_token:
//          type: apiKey
//          name: Authorization
//          in: header
//     oauth2:
//         type: oauth2
//         authorizationUrl: /oauth2/access_token
//         tokenUrl: /oauth2/access_token
//         in: header
//         scopes:
//           'read:profile': allows reading of user profile
//           'write:profile': allows writing of user profile
//         flow: accessCode
//
//     Extensions:
//     x-meta-value: value
//     x-meta-array:
//       - value1
//       - value2
//     x-meta-array-obj:
//       - name: obj
//         value: field
//     x-tagGroups:
//       - name: "User"
//         tags:
//           - account
//           - profile
//           - totp
//           - oauth2
// swagger:meta
package main

import (
	"context"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/handlers"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var db *sqlx.DB
var stage string
var port string
var _RPID string
var _RPOrigin string
var connStr string
var issuerName string

func init() {
	stage = GetEnvWithDefault("STAGE", "localhost")

	port = ":" + GetEnvWithDefault("HTTP_PORT", "8080")

	_RPID = GetEnvWithDefault("RPID", "localhost")

	issuerName = _RPID

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
		users.PATCH("/:id", UpdateUser(userService))
		users.DELETE("/:id", DeleteUser(userService))
	}
	r.POST("/users", Signup(userService))

	// ----- User Authentication ----------------------------------
	totp := r.Group("/totp")
	totp.Use(jwt.TokenValidator(userService))
	{
		totp.GET("/qr-code-raw", NewTOTPRaw(userService, issuerName))
		totp.GET("/qr-code-json", NewTOTPJson(userService, issuerName))
		totp.POST("/confirm", ConfirmToken(userService))
	}

	r.POST("/signin", Signin(userService, jwt.TTL, issuerName))
	r.POST("/refresh", jwt.RefreshToken(userService, issuerName))

	webauthentication := r.Group("/webauthn")
	webauthentication.Use(jwt.TokenValidator(userService))
	{
		webauthentication.POST("/register/begin", BeginRegistration(userService, web))
		webauthentication.POST("/register/finish", FinishRegistration(userService, web))
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
	r.GET("/oauth2/scope/:scope", OAuth2Scope(userService))

	// ---- Client --------------------------------
	r.GET("/oauth2/clients/:id", GetClientName(userService))

	log.Printf("rpid = %s, rporigin = %s \n", _RPID, _RPOrigin)

	switch stage {
	case "localhost":
		log.Println(r.Run(port))
	case "live":
		certFile, keyFile := GetCertAndKey()
		log.Println(r.RunTLS(port, certFile, keyFile))
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

// GetEnvWithDefault attempts to retrieve from env. default calculated based on stage if env value empty.
func GetEnvWithDefault(env, defaultV string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultV
	}
	return v
}
