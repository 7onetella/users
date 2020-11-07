package handlers

import (
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	"github.com/7onetella/users/api/internal/model/oauth2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

// swagger:operation POST /oauth2/authorize oauth2authorize
//
// ---
// summary: "OAuth2 Authorize"
// tags:
//   - oauth2
// parameters:
//   - in: "body"
//     name: "body"
//     description: "Authorization Request"
//     required: true
//     schema:
//       "$ref": "#/definitions/AuthorizationRequest"
// produces:
//   - application/json
// responses:
//   '200':
//     description: replying with authorization code
//     schema:
//       type: object
//       "$ref": "#/definitions/AuthorizationResponse"
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

		response := oauth2.AuthorizationResponse{
			Code:        code,
			RedirectURI: ar.RedirectURI,
			Nonce:       ar.Nonce,
			State:       ar.State,
		}

		c.JSON(200, response)
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
			return
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
		jti := uuid.New().String()
		accessTokenStr, _ := EncodeOAuth2Token(jti, "7onetella", authCode.UserID, clientID, grants.Scope, userFromDB.JWTSecret, 60*60*24)

		// persist the user_id, access token for token validation from resource servers
		accessToken := oauth2.AccessToken{
			TokenID: jti,
			UserID:  authCode.UserID,
			Token:   accessTokenStr,
		}
		dberr = service.StoreAccessTokenForUser(accessToken)
		if rh.HandleDBError(dberr) {
			return
		}

		// if successful then serve the access token
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(200, gin.H{
			"access_token":  accessTokenStr, // jwt token
			"token_type":    "bearer",
			"expires_in":    1 * 60 * 60 * 24,
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

func DoesRedirectURIMatch(redirectURI string) bool {
	return true
}

func OAuth2Scope(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		scope := c.Param("scope")

		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(200, gin.H{
			"scope": scope,
			"desc":  "read user's profile", // random string token
		})

	}
}
