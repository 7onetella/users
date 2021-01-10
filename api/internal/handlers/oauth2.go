package handlers

import (
	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
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
// description: |
//   Handles OAuth2 authorization. Only grant type of `code` has been implemented.
//   Refer to <a href="https://oauth.net/2/">**OAuth2**</a> documentation for details. The OAuth2 `client` signup is
//   missing from UI as well as the User API endpoints. The OAuth2 `client` signup will be added in the future.
//   User is expected to be logged in when this call is made.
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
//     description: authorization code response
//     schema:
//       type: object
//       "$ref": "#/definitions/AuthorizationResponse"
func OAuth2Authorize(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		user, err := r.UserFromContext()
		if err != nil {
			r.AbortWithStatusUnauthorizedError(AuthenticationError, UserUnknown)
			return
		}
		r.WriteCORSHeader()

		ar := oauth2.AuthorizationRequest{}
		err = c.BindJSON(&ar)
		if err != nil {
			r.AbortWithStatusUnauthorizedError(JSONAPISpecError, Unmarshalling)
		}

		// does client exist?
		clientExists, dberr := userService.DoesClientExist(ar.ClientID)
		if dberr != nil {
			r.Logf("oauth2.authorize.failed client_id=%s error=%s", ar.ClientID, dberr)
			r.AbortWithStatusUnauthorizedError(JSONAPISpecError, Unmarshalling)
		}
		// if no then send error
		if !clientExists {
			r.Logf("oauth2.client-check.failed client_id=%s error=%s", ar.ClientID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, ClientNotFound)
			return
		}

		// has this request been made before with the current nonce
		if userService.NonceUsedBefore(ar.ClientID, user.ID, ar.Nonce) {
			r.Logf("oauth2.nonce-check.failed client_id=%s user_id=%s nonce=%s, error=%s", ar.ClientID, user.ID, ar.Nonce, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, OAuth2NonceUsedAlready)
			return
		}

		// update permission for the user
		grant := oauth2.UserGrants{UserID: user.ID, ClientID: ar.ClientID, Scope: ar.Scope}
		dberr = userService.UpdatePermissions(grant)
		if dberr != nil {
			r.Logf("oauth2.updating-permission.failed grant=%#v error=%s", grant, dberr)
			r.AbortWithStatusUnauthorizedError(JSONAPISpecError, Unmarshalling)
		}

		code := uuid.New().String()
		authorizationCode := oauth2.AuthorizationCode{Code: code, ClientID: ar.ClientID, CreatedAt: time.Now().Unix(), UserID: user.ID}
		dberr = userService.StoreAuthorizationRequestCode(authorizationCode)
		if dberr != nil {
			r.Logf("oauth2.updating-permission.failed authorization_code=%#v error=%s", authorizationCode, dberr)
			r.AbortWithStatusUnauthorizedError(JSONAPISpecError, Unmarshalling)
		}

		response := oauth2.AuthorizationResponse{
			Code:        code,
			RedirectURI: ar.RedirectURI,
			Nonce:       ar.Nonce,
			State:       ar.State,
		}

		c.JSON(http.StatusOK, response)
	}
}

// swagger:operation POST /oauth2/access_token oauth2accesstoken
//
// ---
// summary: OAuth2 Access Token
// description: |
//   Exchange code for access token. Access token is in JWT format. The payload looks like the following when it's decoded.
//   ```
//    {
//      "jti":   jti,
//      "iss":   issuer,
//      "sub":   user,
//      "aud":   client,
//      "iat":   1596747095,
//      "exp":   1596747095,
//      "scope": scope
//	  }
//   ```
// consumes:
//   - application/x-www-form-urlencoded
// tags:
//   - oauth2
// parameters:
//   - in: "body"
//     name: "body"
//     description: HTML form containing the following OAuth2 fields
//     required: true
//     schema:
//       type: object
//       properties:
//         grant_type:
//           type: string
//           require: true
//           example: authorization_code
//         code:
//           type: string
//           example: f7cd9875-8386-4d16-97ef-7ae858ebe4c2
//         client_id:
//           type: string
//           example: 352b6e64-e498-4307-b64d-ec9e5b9da65c
//         client_secret:
//           type: string
//           example: 1234567
//         redirect_uri:
//           type: string
//           example: http://www.example.com/oauth2/redirect
// produces:
//   - application/json
// responses:
//   '200':
//     description: OAuth2 access token response
//     schema:
//       type: object
//       "$ref": "#/definitions/AccessTokenResponse"
// security: []
func OAuth2AccessToken(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

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
		authenticated, dberr := userService.AuthenticateClientUsingSecret(clientID, clientSecret)
		if dberr != nil {
			r.Logf("oauth2.access-token-authentication.failed client_id error=%s", clientID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, ValidatingClientFailed)
			return
		}

		if !authenticated {
			r.AbortWithStatusUnauthorizedError(AuthenticationError, ValidatingClientFailed)
			return
		}

		if !DoesRedirectURIMatch(redirectURI) {
			c.JSON(401, gin.H{
				"status": "error",
				"reason": "redirect uri does not match original request",
			})
		}

		// find authorization request record with (code, client_id)
		authCode, dberr := userService.GetAuthorizationRequestCode(code, clientID)
		if dberr != nil {
			r.Logf("oauth2.find-auth-request-code.failed client_id error=%s", clientID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, AuthorizationRequestRecordNotFound)
			return
		}

		// make sure the authorization request was within last 20 seconds
		//if (time.Now().Unix() - authCode.CreatedAt) > 20 {
		//	c.JSON(401, gin.H{
		//		"status": "error",
		//		"reason": "access token request took longer than 20 seconds from authorization request",
		//	})
		//}

		userFromDB, dberr := userService.Get(authCode.UserID)
		if dberr != nil {
			r.Logf("oauth2.finding-user.failed user_id error=%s", authCode.UserID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, UserUnknown)
			return
		}

		grants, dberr := userService.GetUserGrantsForClient(authCode.UserID, clientID)
		if dberr != nil {
			r.Logf("oauth2.finding-grants-for-user.failed user_id client_id=%s error=%s", authCode.UserID, clientID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, UserUnknown)
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
		dberr = userService.StoreAccessTokenForUser(accessToken)
		if dberr != nil {
			r.Logf("oauth2.storing-access-token.failed error=%s", dberr)
			r.AbortWithStatusUnauthorizedError(DatabaseError, PersistingFailed)
			return
		}

		// if successful then serve the access token
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(200, gin.H{
			"access_token":  accessTokenStr, // jwt token
			"token_type":    "Bearer",
			"expires_in":    1 * 60 * 60 * 24,
			"refresh_token": "IwOGYzYTlmM2YxOTQ5MGE3YmNmMDFkNTVk", // random string token
			"scope":         grants.Scope,                         // openid https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile
		})
	}
}

func GetClientName(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()
		clientID := c.Param("id")

		client, dberr := userService.GetClient(clientID)
		if dberr != nil {
			r.Logf("oauth2.finding-client.failed client_id=%s error=%s", clientID, dberr)
			r.AbortWithStatusUnauthorizedError(AuthenticationError, ClientNotFound)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"name": client.Name,
		})
	}
}

func DoesRedirectURIMatch(redirectURI string) bool {
	// doing this to suppress ide warning
	log.Println(redirectURI)
	return true
}

func OAuth2Scope(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c, userService)
		rh.WriteCORSHeader()

		scope := c.Param("scope")

		// ping to suppress ide warning
		_ = userService.Ping()
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(200, gin.H{
			"scope": scope,
			"desc":  "read user's profile", // random string token
		})

	}
}
