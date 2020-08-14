package handlers

import (
	"encoding/base64"
	"github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/xlzd/gotp"
	"log"
	"rsc.io/qr"
	"time"
)

func NewTOTPRaw(userService dbutil.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserFromContext(c)
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		secret := gotp.RandomSecret(16)
		totp := gotp.NewDefaultTOTP(secret)
		url := totp.ProvisioningUri(user.Email, "7onetella")
		log.Println("url = " + url)

		user.TOTPSecretTmp = secret
		user.TOTPSecretTmpExp = dbutil.CurrentTimeInSeconds() + 60*5
		userService.UpdateTOTPTmp(user)

		qrBytes, err := QR(url)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		w := c.Writer
		w.Header().Add("Content-Type", "image/png")
		w.Write(qrBytes)
	}
}

func NewTOTPJson(userService dbutil.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()

		user, err := rh.UserFromContext()
		if rh.HandleError(err) {
			return
		}

		secret := gotp.RandomSecret(16)
		totp := gotp.NewDefaultTOTP(secret)
		url := totp.ProvisioningUri(user.Email, "7onetella")
		log.Println("url = " + url)

		user.TOTPSecretTmp = secret
		user.TOTPSecretTmpExp = dbutil.CurrentTimeInSeconds() + 60*5
		userService.UpdateTOTPTmp(user)

		qrBytes, err := QR(url)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		w := c.Writer
		w.Header().Add("Content-Type", "image/png")
		payload := base64.StdEncoding.EncodeToString(qrBytes)
		c.JSON(200, gin.H{
			"payload": payload,
		})
	}
}

func ConfirmToken(service dbutil.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()
		user, _ := rh.UserFromContext()

		var cred Credentials
		c.ShouldBindJSON(&cred)
		log.Printf("cred json = %#v", cred)

		totp := gotp.NewDefaultTOTP(user.TOTPSecretTmp)

		log.Printf(":token = %s", cred.TOTP)
		now := int(time.Now().Unix())

		log.Printf("timestamp = %d", now)
		verified := totp.Verify(cred.TOTP, now)
		if !verified {
			c.JSON(401, gin.H{
				"status": "invalid",
			})
			c.Abort()
			return
		}

		user.TOTPEnabled = true
		user.TOTPSecretCurrent = user.TOTPSecretTmp
		user.TOTPSecretTmp = ""
		user.TOTPSecretTmpExp = 0
		log.Printf("user from context = \n%#v", user)
		dberr := service.UpdateTOTP(user)
		if rh.HandleDBError(dberr) {
			c.JSON(401, gin.H{
				"status": "invalid",
			})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{
			"status": "valid",
		})
	}
}

func isTOTPValid(user *User, token string) bool {
	totp := gotp.NewDefaultTOTP(user.TOTPSecretCurrent)
	log.Printf(":token = %s", token)
	now := int(time.Now().Unix())
	log.Printf("timestamp = %d", now)
	return totp.Verify(token, now)
}

func UserFromContext(c *gin.Context) User {
	ctx := c.Request.Context()
	user, _ := ctx.Value("user").(User)
	return user
}

func QR(url string) ([]byte, error) {
	code, err := qr.Encode(url, qr.Q)
	if err != nil {
		return nil, err
	}
	return code.PNG(), nil
}
