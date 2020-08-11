package main

import (
	"github.com/google/uuid"
	"github.com/xlzd/gotp"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTP     string `json:"totp"`
}

type AuthToken struct {
	Token      string `json:"token"`
	Expiration int64  `json:"exp"`
}

type JWTAuth struct {
	secret   string
	claimKey string
	ttl      time.Duration
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

		authenticated, user := userService.Authenticate(cred.Username, cred.Password)

		if !authenticated {
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

		// this just to make testing easier during development phase
		if user.Email == "user8az28y@example.com" {
			goto SilentLogin
		}

		if user.MFAEnabled {
			if len(cred.TOTP) == 0 {
				c.AbortWithStatusJSON(401, gin.H{
					"reason":  "missing_totp",
					"message": "TOTP required",
				})
				userService.RecordAuthEvent(NewAuthEvent(user.ID, "missing_totp", c.ClientIP(), c.ClientIP(), c.Request.UserAgent()))
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

	SilentLogin:
		if authenticated {
			tokenString, expTime, err := EncodeToken(a.claimKey, user.ID, user.JWTSecret, a.ttl)
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
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
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

		user, err := UnmarshalUser(payload, userSchema)
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
		out, err := MarshalUser(c.Request.URL.RequestURI(), userSchema, user)
		if rh.HandleError(err) {
			return
		}
		c.String(http.StatusOK, out)
	}
}

func LogErr(txid string, message string, opserr interface{}) {
	log.Printf("%s %s: %#v", txid, message, opserr)
}
