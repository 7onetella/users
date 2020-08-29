package handlers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
	"time"
)

type CustomClaims struct {
	jwt.StandardClaims
	Scope string `json:"scope"`
}

func (cc CustomClaims) Valid() error {
	log.Println("standard claims valid()")
	return cc.StandardClaims.Valid()
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

// DecodeToken decodes jwt token then returns claims in map
func DecodeTokenAsCustomClaims(tokenString, secret string) (CustomClaims, error) {
	cm, err := DecodeToken(tokenString, secret)
	if err != nil {
		return CustomClaims{}, err
	}

	toStr := func(m map[string]interface{}, attr string) string {
		if s, ok := m[attr].(string); ok {
			return s
		}
		return ""
	}

	toInt64 := func(m map[string]interface{}, attr string) int64 {
		if f, ok := m[attr].(float64); ok {
			return int64(f)
		}
		return 0
	}

	claims := CustomClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  toStr(cm, "aud"),
			ExpiresAt: toInt64(cm, "exp"),
			Id:        toStr(cm, "jti"),
			IssuedAt:  toInt64(cm, "iat"),
			Issuer:    toStr(cm, "iss"),
			NotBefore: toInt64(cm, "nbf"),
			Subject:   toStr(cm, "sub"),
		},
		Scope: toStr(cm, "scope"),
	}

	return claims, nil
}

// EncodeID signs id
func EncodeToken(userID, secret string, ttl time.Duration) (string, time.Time, error) {
	expTime := time.Now().Add(ttl * time.Second)

	claim := jwt.MapClaims{
		"iss": "7onetella",
		"sub": userID,
		"aud": "7onetella",
		"iat": time.Now().Unix(),
		"exp": expTime.Unix(),
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))

	return tokenString, expTime, err
}

func ParseAuthTokenFromAuthorizationHeader(req *http.Request) (string, error) {
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
func EncodeOAuth2Token(jti, issuer, user, client, scope, secret string, ttl time.Duration) (string, error) {
	expTime := time.Now().Add(ttl * time.Second)

	claim := jwt.MapClaims{
		"jti":   jti,
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
