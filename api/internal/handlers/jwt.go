package handlers

import (
	"context"
	"encoding/json"
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/7onetella/users/api/pkg/jwtutil"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/xlzd/gotp"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
	EventID  string `json:"event_id"`
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
			w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.WriteHeader(http.StatusAccepted)
			c.Abort()
			return
		}
		c.Next()
	}
}

// Signin signs in user
func (a JWTAuth) Signin(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		cred := Credentials{}
		c.ShouldBind(&cred)

		var user *User
		var dberr *DBOpError

		if len(cred.EventID) > 0 {
			user, dberr = userService.FindByEventID(cred.EventID)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_event_id",
					"message": "TOTP auth failed",
				})
				return
			}
			goto CheckTOTP
		}

		if len(cred.Username) > 0 {
			// this just to make testing easier during development phase
			user, dberr = userService.FindByEmail(cred.Username)
			if dberr != nil {
				log.Printf("error while authenticating: %v", dberr)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "server_error",
					"message": "Authentication failed",
				})
			}

			if user.Password != cred.Password {
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_credential",
					"message": "Check login name or password",
				})
				dberr := userService.RecordAuthEvent(NewAuthEvent(user.ID, "invalid_credential", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
				if rh.HandleDBError(dberr) {
					return
				}
				return
			}

			if user.Email == "user8az28y@example.com" {
				goto GrantAccess
			}
		}

	CheckTOTP:
		if user.TOTPEnabled {
			if len(cred.TOTP) == 0 {
				event := NewAuthEvent(user.ID, "missing_totp", c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
				userService.RecordAuthEvent(event)
				c.AbortWithStatusJSON(401, gin.H{
					"reason":   "missing_totp",
					"message":  "TOTP required",
					"event_id": event.ID,
				})
				return
			}
			if !isTOTPValid(user, cred.TOTP) {
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "invalid_totp",
					"message": "Check your TOTP",
				})
				userService.RecordAuthEvent(NewAuthEvent(user.ID, "invalid_totp", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
				return
			}
		}

	GrantAccess:
		tokenString, expTime, err := jwtutil.EncodeToken(a.ClaimKey, user.ID, user.JWTSecret, a.TTL)
		if err != nil {
			log.Println("encoding error")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		log.Println("Signin successful")
		//log.Println("Sign-In successful dropping token cookie")
		//http.SetCookie(w, &http.Cookie{
		//	Name:    "token",
		//	Value:   tokenString,
		//	Expires: expTime,
		//})

		token := AuthToken{
			Token:      tokenString,
			Expiration: expTime.Unix(),
		}
		c.JSON(200, token)
		userService.RecordAuthEvent(NewAuthEvent(user.ID, "successful_login", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
	}
}

// Signup signs up user
func Signup(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)

		rh.WriteCORSHeader()

		payload, errs := rh.GetPayload(User{})
		if rh.HandleError(errs...) {
			return
		}

		user, err := UnmarshalUser(payload, UserJSONSchema)
		if rh.HandleError(err) {
			return
		}

		user.ID = uuid.New().String()
		user.Created = CurrentTimeInSeconds()
		user.PlatformName = "web"
		user.JWTSecret = gotp.RandomSecret(16)

		dberr := userService.Register(user)
		if rh.HandleDBError(dberr) {
			return
		}

		rh.SetContentTypeJSON()
		out, err := MarshalUser(c.Request.URL.RequestURI(), UserJSONSchema, user)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, out)
	}
}

func (a JWTAuth) Validator(service UserService) gin.HandlerFunc {
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
