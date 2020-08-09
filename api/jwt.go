package main

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

func (a JWTAuth) Validator(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := parseAuthTokenFromHeader(c.Request)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		log.Printf("Authentication: Bearer %s", tokenString)

		claims, err := DecodeToken(tokenString, a.secret)
		if err != nil {
			log.Printf("error decoding: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if userID, ok := claims[a.claimKey].(string); ok {
			req := c.Request
			user, operr := userService.Get(userID)
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

func parseAuthTokenFromHeader(req *http.Request) (string, error) {
	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		return "", errors.New("header Authorization is empty")
	}
	// strip out Bearer from value
	terms := strings.Split(authorization, " ")
	token := terms[1]
	return token, nil
}

// DecodeToken decodes jwt token then returns claims in map
func DecodeToken(tokenString, secret string) (map[string]interface{}, error) {

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	log.Printf("claim = %#v", token.Claims)

	if mc, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		claims := map[string]interface{}(mc)
		return claims, nil
	}

	return nil, errors.New("token decoding error")

}

// EncodeID signs id
func EncodeToken(claimKey, claimValue, secret string, ttl time.Duration) (string, time.Time, error) {
	expTime := time.Now().Add(ttl * time.Second)
	claim := jwt.MapClaims{
		claimKey: claimValue,
		"exp":    expTime.Unix(),
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))

	return tokenString, expTime, err
}

// SigninHandlerFunc signs in user
func (a JWTAuth) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Token Refresh Started")
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		var rt AuthToken
		c.ShouldBindJSON(&rt)

		log.Println("token =", rt.Token)

		claims, err := DecodeToken(rt.Token, a.secret)
		if err != nil {
			log.Println("Issue with decoding", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claimValue, _ := claims[a.claimKey].(string)
		tokenString, expTime, err := EncodeToken(a.claimKey, claimValue, a.secret, 120)
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
