package handlers

import (
	"encoding/base64"
	"log"
	"time"

	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/xlzd/gotp"
	"rsc.io/qr"
)

// swagger:operation GET /totp/qr-code-raw qrcoderaw
//
// ---
// summary: "Generate new QR code image PNG"
// tags:
//   - totp
// produces:
//   - application/json
// responses:
//   '200':
//     description: QR code image in PNG format
//     content:
//       image/png:
//         schema:
//           type: string
//           format: binary
func NewTOTPRaw(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		user, err := rh.UserFromContext()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		rh.WriteCORSHeader()

		secret := gotp.RandomSecret(16)
		totp := gotp.NewDefaultTOTP(secret)
		url := totp.ProvisioningUri(user.Email, "7onetella")
		log.Println("url = " + url)

		user.TOTPSecretTmp = secret
		user.TOTPSecretTmpExp = CurrentTimeInSeconds() + 60*5
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

// swagger:operation GET /totp/qr-code-json qrcodejson
//
// ---
// summary: "Generate new QR code image JSON"
// tags:
//   - totp
// produces:
//   - application/json
// responses:
//   '200':
//     description: QR code image in encoded in JSON format
//     schema:
//       type: object
//       properties:
//         payload:
//           type: string
//           description: PNG image encoded in base64
func NewTOTPJson(userService UserService) gin.HandlerFunc {
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
		user.TOTPSecretTmpExp = CurrentTimeInSeconds() + 60*5
		userService.UpdateTOTPTmp(user)

		qrBytes, err := QR(url)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		w := c.Writer
		w.Header().Add("Content-Type", "image/png")
		payload := base64.StdEncoding.EncodeToString(qrBytes)
		// header used for validating totp in testing
		c.Header("x-totp", totp.Now())
		c.JSON(200, gin.H{
			"payload": payload,
		})
	}
}

// swagger:operation POST /totp/confirm confirm
//
// ---
// summary: "Confirm QR Code"
// tags:
//   - totp
// parameters:
//   - in: body
//     name: totp
//     description: TOTP code
//     schema:
//       type: object
//       required:
//         - totp
//       properties:
//         totp:
//           type: string
// produces:
//   - application/json
// responses:
//   '200':
//     description: confirmation successful
//     schema:
//       type: object
//       properties:
//         status:
//           type: string
//           description: confirmation status
//           example: totp enabled
//   '401':
//     description: confirmation successful
//     schema:
//       type: object
//       properties:
//         status:
//           type: string
//           description: confirmation status
//           example: invalid
func ConfirmToken(service UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rh := NewRequestHandler(c)
		rh.WriteCORSHeader()
		user, _ := rh.UserFromContext()

		var cred TOTPCredentials
		c.ShouldBindJSON(&cred)
		log.Printf("cred json = %#v", cred)

		totp := gotp.NewDefaultTOTP(user.TOTPSecretTmp)

		log.Printf("totp = %s", cred.TOTP)
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
			"status": "totp enabled",
		})
	}
}

func IsTOTPValid(user User, token string) bool {
	totp := gotp.NewDefaultTOTP(user.TOTPSecretCurrent)
	log.Printf(":token = %s", token)
	now := int(time.Now().Unix())
	log.Printf("timestamp = %d", now)
	return totp.Verify(token, now)
}

func QR(url string) ([]byte, error) {
	code, err := qr.Encode(url, qr.M)
	if err != nil {
		return nil, err
	}
	return code.PNG(), nil
}
