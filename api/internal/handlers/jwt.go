package handlers

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/dgrijalva/jwt-go"

	"net/http"

	"github.com/gin-gonic/gin"
)

// CredentialsBase represents user credentials type
//
// swagger:model
type CredentialsBase struct {
	//
	// discriminator: true
	Type string `json:"type"`
}

// This struct combines all the attributes from PasswordCredentials, TOTPCredentials and WebauthnCredentials.
// These three are used for generating swagger documentation. This struct is the real struct used internally during
// authentication.
type Credentials struct {
	// this field
	Type string `json:"type"`

	Username string `json:"username"`

	Password string `json:"password"`

	SigninSessionToken string `json:"signin_session_token"`

	TOTP string `json:"totp"`

	WebauthnAuthToken string `json:"webauthn_session_token"`
}

// password_credentials represents user password credentials
//
// swagger:model password_credentials
type PasswordCredentials struct {
	// swagger:allOf com.7onetella.PasswordCredentials
	CredentialsBase
	// username
	//
	// required: true
	// example: john.smith@example.com
	Username string `json:"username"`
	// password
	//
	// required: true
	// example: password1234
	Password string `json:"password"`
}

// totp_credentials represents user totp credentials
//
// swagger:model totp_credentials
type TOTPCredentials struct {
	// swagger:allOf com.7onetella.TOTPCredentials
	CredentialsBase
	// token received after successful password authentication
	//
	// required: true
	// example: MzM4OGNkMWEtNmQyNC00MDQ1LWJmYzctMWJlMzM3ZTk1NDQ5
	SigninSessionToken string `json:"signin_session_token"`
	// totp code from authenticator device
	//
	// required: true
	// example: 327621
	TOTP string `json:"totp"`
}

// webauthn_credentials represents user webauthn credentials
//
// swagger:model webauthn_credentials
type WebauthnCredentials struct {
	// swagger:allOf com.7onetella.WebauthnCredentials
	CredentialsBase
	// token received after successful password authentication
	//
	// required: true
	// example: MzM4OGNkMWEtNmQyNC00MDQ1LWJmYzctMWJlMzM3ZTk1NDQ5
	SigninSessionToken string `json:"signin_session_token"`
	// token received after successful webauthn authentication
	//
	// required: true
	// example: ZDkyZjJkNWMtNGU2Ny00ZGRmLWI2ZGQtOTExNTcyYzIwNWFk
	WebauthnAuthToken string `json:"webauthn_session_token"`
}

// This is JWT Auth Token
//
// swagger:model authToken
type AuthToken struct {
	// jwt token
	// required: true
	Token string `json:"token"`
	// token expiration date
	// required: true
	Expiration int64 `json:"exp"`
}

type JWTAuth struct {
	ClaimKey string
	TTL      time.Duration
}

func (c Credentials) IsSigninSessionTokenPresent() bool {
	return len(c.SigninSessionToken) > 0
}

func (c Credentials) IsUsernamePresent() bool {
	return len(c.Username) > 0
}

func (c Credentials) IsTOTPCodePresent() bool {
	return len(c.TOTP) > 0
}

func (c Credentials) IsWebauthnTokenPresent() bool {
	return len(c.WebauthnAuthToken) > 0
}

// PreflightOptionsHandler handles preflight OPTIONS
func PreflightOptions() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			w := c.Writer
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS, HEAD, POST, PUT, PATCH, DELETE")
			w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, AuthToken")
			w.WriteHeader(http.StatusAccepted)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a JWTAuth) TokenValidator(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := ParseAuthTokenFromAuthorizationHeader(c.Request)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		log.Printf("Authentication: Bearer %s", tokenString)
		terms := strings.Split(tokenString, ".")
		payloadRaw := terms[1]
		jwtSecret, dberr := GetJWTSecretForUser(payloadRaw, service)
		if dberr != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		claims, err := DecodeTokenAsCustomClaims(tokenString, jwtSecret)
		if err != nil {
			log.Printf("error decoding: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// if this is oauth2 delegated api call
		// make sure the token was actually issued
		if claims.Issuer != claims.Audience {
			accessToken, dberr := service.GetAccessToken(claims.Id)
			if dberr != nil {
				log.Printf("token not found")
				c.AbortWithError(http.StatusUnauthorized, dberr.Err)
				return
			}
			if accessToken.Token != tokenString {
				log.Printf("invalid access token")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if claims.Subject != accessToken.UserID {
				log.Printf("user id differs from what's stored in db")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		userID := claims.Subject
		req := c.Request
		user, operr := service.Get(userID)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, operr.Err)
			return
		}
		ctx := context.WithValue(req.Context(), "user", user)
		c.Request = req.Clone(ctx)
		c.Next()
	}
}

func GetJWTSecretForUser(payloadRaw string, userService UserService) (string, *DBOpError) {
	log.Printf("payload = %s", payloadRaw)
	claims, err := ExtractClaimsFromPayload(payloadRaw)
	if err != nil {
		log.Printf("jmap err = %v", err)
	}
	log.Printf("jmap = %#v", claims)
	claimedUser, dberr := userService.Get(claims.Subject)
	if dberr != nil {
		log.Print(err)
		return "", dberr
	}
	return claimedUser.JWTSecret, nil
}

func ExtractClaimsFromPayload(s string) (*CustomClaims, error) {
	data, err := jwt.DecodeSegment(s)
	if err != nil {
		return nil, err
	}

	claims := &CustomClaims{}
	err = json.Unmarshal(data, claims)

	return claims, err
}

// SigninHandlerFunc signs in user
func (a JWTAuth) RefreshToken(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Token Refresh Started")
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		var rt AuthToken
		c.ShouldBindJSON(&rt)

		terms := strings.Split(rt.Token, ".")
		payloadRaw := terms[1]
		jwtSecret, dberr := GetJWTSecretForUser(payloadRaw, service)
		if dberr != nil {
			c.AbortWithError(http.StatusUnauthorized, dberr.Err)
			return
		}

		log.Println("token =", rt.Token)

		claims, err := DecodeTokenAsCustomClaims(rt.Token, jwtSecret)
		if err != nil {
			log.Println("Issue with decoding", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID := claims.Subject
		tokenString, expTime, err := EncodeToken(userID, jwtSecret, 120)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		log.Println("Token refresh successful")

		refreshToken := AuthToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}

		c.JSON(200, refreshToken)
		//data, err := json.Marshal(&refreshToken)
		//if err != nil {
		//	c.AbortWithError(500, err)
		//	return
		//}
		//w.Write(data)

	}
}
