package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	. "github.com/7onetella/users/api/internal/dbutil"
	. "github.com/7onetella/users/api/internal/httputil"
	. "github.com/7onetella/users/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/xlzd/gotp"
	"rsc.io/qr"
)

// Hiding endpoint to not to confuse developer reading the API
func NewTOTPRaw(userService UserService, totpIssuerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		user, err := r.UserFromContext()
		if err != nil {
			r.DenyAccessForAnonymous(AuthenticationError, SigninSessionTokenDecodingFailed)
			return
		}
		r.WriteCORSHeader()

		secret := gotp.RandomSecret(16)
		totp := gotp.NewDefaultTOTP(secret)
		url := totp.ProvisioningUri(user.Email, totpIssuerName)

		user.TOTPSecretTmp = secret
		user.TOTPSecretTmpExp = CurrentTimeInSeconds() + 60*5
		dberr := userService.UpdateTOTPTmp(user)
		if dberr != nil {
			dberr.Log(r.TX())
			c.AbortWithStatusJSON(http.StatusBadRequest, New(DatabaseError, PersistingFailed))
			return
		}

		qrBytes, err := QR(url)
		if err != nil {
			LogErr(r.TX(), "error encoding qr code", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(TOTPError, ProblemEncodingQRCode))
			return
		}

		w := c.Writer
		w.Header().Add(ContentType, ImagePNG)
		_, err = w.Write(qrBytes)
		if err != nil {
			LogErr(r.TX(), "error writing image bytes", err)
		}
	}
}

// swagger:operation GET /totp/qr-code-json qrcodejson
//
// ---
// summary: "New TOTP QR code"
// description: |
//   New TOTP QR code image is generated in PNG format. The image is then encoded in BASE64 and returned back to the client
//   in JSON response. The new TOTP QR code won't be associated with the user's account unless `/totp/confirm` endpoint is provided with
//   the TOTP code from the authenticator device. Refer to <a href="/accounts/redoc.html#operation/confirm">**Confirm TOTP QR Code**</a>
//   documentation for details.
// tags:
//   - totp
// produces:
//   - application/json
// responses:
//   '200':
//     description: QR code image PNG encoded in base64
//     schema:
//       type: object
//       properties:
//         payload:
//           type: string
//           description: PNG image encoded in base64
//   '401':
//     description: problem with retrieving user from authorization header
//   '500':
//     description: internal server error
func NewTOTPJson(userService UserService, totpIssuerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()

		user, err := r.UserFromContext()
		if err != nil {
			r.DenyAccessForAnonymous(AuthenticationError, UserUnknown)
			return
		}

		secret := gotp.RandomSecret(32)
		totp := gotp.NewTOTP(secret, 6, 30, nil)
		url := totp.ProvisioningUri(user.Email, totpIssuerName)

		// persist the totp secret and expiration time of 5 minutes
		user.TOTPSecretTmp = secret
		user.TOTPSecretTmpExp = CurrentTimeInSeconds() + 60*5
		dberr := userService.UpdateTOTPTmp(user)
		if dberr != nil {
			dberr.Log(r.TX())
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(DatabaseError, PersistingFailed))
			return
		}

		qrBytes, err := QR(url)
		if err != nil {
			LogErr(r.TX(), "error encoding qr code", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(TOTPError, ProblemEncodingQRCode))
			return
		}

		w := c.Writer
		w.Header().Add(ContentType, ImagePNG)
		payload := base64.StdEncoding.EncodeToString(qrBytes)
		// header used for validating totp in testing
		// only authorized users can access their own x-totp header
		c.Header(XTOTP, totp.Now())
		c.JSON(http.StatusOK, gin.H{
			"payload": payload,
		})
	}
}

// swagger:operation POST /totp/confirm confirm
//
// ---
// summary: Confirm TOTP QR Code
// description: |
//   Confirming QR code is part of enabling TOTP process. User will be shown QR code on screen and will be asked to scan the QR
//   code using his/her authenticator. Entering the TOTP code shown on the authenticator confirms that the device's time
//   is in sync. Entering also confirms that TOTP calculation algorithm on both authenticator and server work as expected.
//   Foregoing confirmation means user's TOTP can be rejected during authentication. This also ensures that user does
//   have access to an working authenticator.
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
//           example: 467292
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
//   '400':
//     description: confirmation failed
//     schema:
//       type: object
//       properties:
//         status:
//           type: string
//           description: confirmation status
//           example: totp invalid
//   '401':
//     description: problem with retrieving user from authorization header
//   '500':
//     description: internal server error
func ConfirmToken(userService UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := NewRequestHandler(c, userService)
		r.WriteCORSHeader()
		user, err := r.UserFromContext()
		if err != nil {
			r.DenyAccessForAnonymous(AuthenticationError, UserUnknown)
			return
		}

		var cred TOTPCredentials
		err = c.ShouldBindJSON(&cred)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(JSONAPISpecError, Marshalling))
			return
		}

		totp := gotp.NewDefaultTOTP(user.TOTPSecretTmp)
		now := int(time.Now().Unix())

		verified := totp.Verify(cred.TOTP, now)
		if !verified {
			c.AbortWithStatusJSON(http.StatusBadRequest, New(TOTPError, InvalidTOTP))
			return
		}

		user.TOTPEnabled = true
		user.TOTPSecretCurrent = user.TOTPSecretTmp
		user.TOTPSecretTmp = ""
		user.TOTPSecretTmpExp = 0

		dberr := userService.UpdateTOTP(user)
		if dberr != nil {
			dberr.Log(r.TX())
			c.AbortWithStatusJSON(http.StatusInternalServerError, New(DatabaseError, PersistingFailed))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "totp enabled",
		})
	}
}

func IsTOTPValid(user User, token string) bool {
	totp := gotp.NewDefaultTOTP(user.TOTPSecretCurrent)
	//log.Printf(":token = %s", token)
	now := int(time.Now().Unix())
	//log.Printf("timestamp = %d", now)
	return totp.Verify(token, now)
}

func QR(url string) ([]byte, error) {
	code, err := qr.Encode(url, qr.M)
	if err != nil {
		return nil, err
	}
	return code.PNG(), nil
}
