package handlers

import (
	"context"
	"encoding/json"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/7onetella/users/api/pkg/jwtutil"
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
		tokenString, err := jwtutil.ParseAuthTokenFromHeader(c.Request)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		log.Printf("Authentication: Bearer %s", tokenString)
		terms := strings.Split(tokenString, ".")
		jwtSecret, dberr := GetJWTSecretFromClaim(terms, service)
		if dberr != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		claims, err := jwtutil.DecodeToken(tokenString, jwtSecret)
		if err != nil {
			log.Printf("error decoding: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if userID, ok := claims[a.ClaimKey].(string); ok {
			req := c.Request
			user, operr := service.Get(userID)
			if err != nil {
				c.AbortWithError(http.StatusUnauthorized, operr.Err)
				return
			}
			ctx := context.WithValue(req.Context(), "user", user)
			c.Request = req.Clone(ctx)
			c.Next()
		} else {
			c.AbortWithError(http.StatusUnauthorized, err)
		}

	}
}

func GetJWTSecretFromClaim(terms []string, userService UserService) (string, *DBOpError) {
	payloadRaw := terms[1]
	log.Printf("payload = %s", payloadRaw)
	claim, err := ExtractUserClaim(payloadRaw)
	if err != nil {
		log.Printf("jmap err = %v", err)
	}
	log.Printf("jmap = %#v", claim)
	claimedUser, dberr := userService.Get(claim.Id)
	if dberr != nil {
		log.Print(err)
		return "", dberr
	}
	return claimedUser.JWTSecret, nil
}

type UserClaims struct {
	ExpiresAt int64  `json:"exp,omitempty"`
	Id        string `json:"user_id,omitempty"`
}

func ExtractUserClaim(s string) (*UserClaims, error) {
	data, err := jwt.DecodeSegment(s)
	if err != nil {
		return nil, err
	}

	claim := &UserClaims{}
	err = json.Unmarshal(data, claim)

	return claim, err
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
		jwtSecret, dberr := GetJWTSecretFromClaim(terms, service)
		if dberr != nil {
			c.AbortWithError(http.StatusUnauthorized, dberr.Err)
			return
		}

		log.Println("token =", rt.Token)

		claims, err := jwtutil.DecodeToken(rt.Token, jwtSecret)
		if err != nil {
			log.Println("Issue with decoding", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claimValue, _ := claims[a.ClaimKey].(string)
		tokenString, expTime, err := jwtutil.EncodeToken(a.ClaimKey, claimValue, jwtSecret, 120)
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
