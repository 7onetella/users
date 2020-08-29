package handlers

import (
	"context"
	"encoding/json"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
)

type Credentials struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	TOTP               string `json:"totp"`
	PrimaryAuthToken   string `json:"auth_token"`
	SecondaryAuthToken string `json:"sec_auth_token"`
}

type AuthToken struct {
	Token      string `json:"token"`
	Expiration int64  `json:"exp"`
}

type JWTAuth struct {
	ClaimKey string
	TTL      time.Duration
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
