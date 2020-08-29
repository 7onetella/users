package jwtutil

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
	"time"
)

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

	//claim := jwt.StandardClaims{
	//	Id:        claimValue,
	//	ExpiresAt: expTime.Unix(),
	//}
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

func ParseAuthTokenFromHeader(req *http.Request) (string, error) {
	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		return "", errors.New("header Authorization is empty")
	}
	// strip out Bearer from value
	terms := strings.Split(authorization, " ")
	token := terms[1]
	return token, nil
}

// EncodeID signs id
func EncodeOAuth2Token(issuer, user, client, scope, secret string, ttl time.Duration) (string, error) {
	expTime := time.Now().Add(ttl * time.Second)

	claim := jwt.MapClaims{
		"iss":   issuer,
		"sub":   user,
		"aud":   client,
		"iat":   time.Now().Unix(),
		"exp":   expTime.Unix(),
		"scope": scope,
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))

	return tokenString, err
}
