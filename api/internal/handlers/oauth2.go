package handlers

import (
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/7onetella/users/api/internal/model/oauth2"
	"github.com/7onetella/users/api/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

func OAuth2Authorize(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		user, err := rh.UserFromContext()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		rh.WriteCORSHeader()

		ar := oauth2.AuthorizationRequest{}
		c.BindJSON(&ar)

		// does client exist?
		clientExists, dberr := service.DoesClientExist(ar.ClientID)
		if rh.HandleDBError(dberr) {
			return
		}
		// if no then send error
		if !clientExists {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "unknown client",
			})
			return
		}

		// has this request been made before with the current nonce
		if service.NonceUsedBefore(ar.ClientID, user.ID, ar.Nonce) {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "nonce has been used already",
			})
			return
		}

		// NamedExec permission for the user
		dberr = service.UpdatePermissions(oauth2.UserGrants{user.ID, ar.ClientID, ar.Scope})
		if rh.HandleDBError(dberr) {
			return
		}

		code := uuid.New().String()
		authorizationCode := oauth2.AuthorizationCode{code, ar.ClientID, time.Now().Unix(), user.ID}
		dberr = service.StoreAuthorizationRequestCode(authorizationCode)
		if rh.HandleDBError(dberr) {
			return
		}

		log.Printf("payload = \n%#v\n", ar)

		c.JSON(200, gin.H{
			"code":         code,
			"redirect_uri": ar.RedirectURI,
			"nonce":        ar.Nonce,
			"state":        ar.State,
		})
	}
}

func OAuth2AccessToken(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		grantType := c.PostForm("grant_type")
		if grantType != "authorization_code" {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "grant type is not authorization_code",
			})
		}

		code := c.PostForm("code")
		clientID := c.PostForm("client_id")
		clientSecret := c.PostForm("client_secret")
		redirectURI := c.PostForm("redirect_uri")

		// is this valid request from client?
		// validate client with client_secret
		authenticated, dberr := service.AuthenticateClientUsingSecret(clientID, clientSecret)
		if rh.HandleDBError(dberr) {
			return
		}

		if !authenticated {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "we don't know who you are",
			})
		}

		if !DoesRedirectURIMatch(redirectURI) {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "redirect uri does not match original request",
			})
		}

		// find authorization request record with (code, client_id)
		authCode, dberr := service.GetAuthorizationRequestCode(code, clientID)
		if rh.HandleDBError(dberr) {
			return
		}

		// make sure the authorization request was within last 20 seconds
		//if (time.Now().Unix() - authCode.CreatedAt) > 20 {
		//	c.JSON(401, gin.H{
		//		"status": "error",
		//		"reason": "access token request took longer than 20 seconds from authorization request",
		//	})
		//}

		userFromDB, dberr := service.Get(authCode.UserID)
		if rh.HandleDBError(dberr) {
			return
		}

		grants, dberr := service.GetUserGrantsForClient(authCode.UserID, clientID)
		if rh.HandleDBError(dberr) {
			return
		}

		// generate the access token
		accessToken, _ := jwtutil.EncodeOAuth2Token("7onetella", authCode.UserID, clientID, grants.Scope, userFromDB.JWTSecret, 60*60*24*365)

		// persist the user_id, access token for token validation from resource servers
		StoreAccessTokenForUser(accessToken, authCode.UserID)

		// if successful then serve the access token
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(200, gin.H{
			"access_token":  accessToken, // jwt token
			"token_type":    "bearer",
			"expires_in":    1 * 60 * 60 * 24 * 365,
			"refresh_token": "IwOGYzYTlmM2YxOTQ5MGE3YmNmMDFkNTVk", // random string token
			"scope":         grants.Scope,
		})
	}
}

func GetClientName(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()
		clientID := c.Param("id")

		client, dberr := service.GetClient(clientID)
		if dberr != nil {
			c.AbortWithError(http.StatusInternalServerError, dberr.Err)
			return
		}

		c.JSON(200, gin.H{
			"name": client.Name,
		})
	}
}

func StoreAccessTokenForUser(accessToken, userID string) {

}

func DoesRedirectURIMatch(redirectURI string) bool {
	return true
}
