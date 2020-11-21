package model

import (
	"encoding/json"
	"fmt"
)

type JSONAPIErrors struct {
	Errors []JSONAPIError `json:"errors"`
}

type JSONAPIError struct {
	Meta *Error `json:"meta,omitempty"`
}

// CREDIT goes to cloudflare https://github.com/cloudflare/cfssl Error

// This is JWT Auth Token
//
// swagger:model AuthError
type Error struct {
	// error code for machine
	ErrorCode int `json:"code"`
	// message to display for end user
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// This is JWT Auth Token
//
// swagger:model MissingDataError
type MissingDataError struct {
	Error
	// token sent after successful password signin
	// example: MzM4OGNkMWEtNmQyNC00MDQ1LWJmYzctMWJlMzM3ZTk1NDQ5
	SigninSessionToken string `json:"signin_session_token"`
}

// Category is the most significant digit of the error code.
type Category int

// Reason is the last 3 digits of the error code.
type Reason int

const (
	// Success indicates no error occurred.
	Success Category = 1000 * iota // 0XXX

	// DatabaseError indicates a fault in database operation.
	DatabaseError // 1XXX

	SecurityError // 2XXX

	JSONAPISpecError // 3XXX

	AuthenticationError // 4xxx

	ServerError // 5xxx

	TOTPError // 6xxx
)

// DatabaseError reasons
const (
	// QueryingFailed indicates general database error during SQL execution
	QueryingFailed Reason = 100 * (iota + 1) //11XX

	PersistingFailed

	Unknown

	GeneralError
)

// None is a non-specified error.
const (
	None Reason = iota
)

// SecurityError reasons
const (
	// QueryingFailed indicates general database error during SQL execution
	ContextUserDoesNotMatchGivenUserID Reason = 100 * (iota + 1) //21XX
)

// JsonError reasons
const (
	// Marshalling indicates general database error during SQL execution
	Marshalling Reason = 100 * (iota + 1) //31XX

	Unmarshalling
)

// AuthenticationError reasons
const (
	// SessionTokenDecodingFailed indicates base64 decoding failure
	SigninSessionTokenDecodingFailed Reason = 100 * (iota + 1) //41XX

	SigninSessionExpired

	UsernameOrPasswordDoesNotMatch

	WebauthnAuthFailure

	TOTPAuthFailure

	UserUnknown

	JWTEncodingFailure

	TOTPRequired

	WebAuthnRequired

	WebauthnRegistrationFailure
)

// ServerError reasons
const (
	QRCodeFailure Reason = 100 * (iota + 1) // 51XX

	RetrievingPayloadError
)

// TOTPError reasons
const (
	InvalidTOTP Reason = 100 * (iota + 1) // 61XX

	ProblemEncodingQRCode
)

func New(category Category, reason Reason) *Error {
	errorCode := int(category) + int(reason)
	var msg string

	switch category {
	case DatabaseError:
		switch reason {
		case QueryingFailed:
			msg = "Database query failed"
		case PersistingFailed:
			msg = "Writing to database failed"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category DatabaseError.",
				reason))
		}

	case SecurityError:
		switch reason {
		case Unknown:
			msg = "Unknown security error"
		case ContextUserDoesNotMatchGivenUserID:
			msg = "User ID in context does not match given user's ID"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category SecurityError.",
				reason))
		}

	case JSONAPISpecError:
		switch reason {
		case Marshalling:
			msg = "Error occurred during marshalling"
		case Unmarshalling:
			msg = "Error occurred during unmarshalling"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category JSONAPISpecError.",
				reason))
		}

	case AuthenticationError:
		switch reason {
		case SigninSessionTokenDecodingFailed:
			msg = "Error decoding signin_session_token"
		case SigninSessionExpired:
			msg = "Your signin session timed out"
		case UsernameOrPasswordDoesNotMatch:
			msg = "Check your username or password"
		case WebauthnAuthFailure:
			msg = "Webauthn authentication failed"
		case TOTPAuthFailure:
			msg = "Your TOTP code is invalid"
		case UserUnknown:
			msg = "User is unknown"
		case JWTEncodingFailure:
			msg = "Problem with encoding jwt token"
		case TOTPRequired:
			msg = "TOTP required"
		case WebAuthnRequired:
			msg = "WebAuthn Auth Required"
		case WebauthnRegistrationFailure:
			msg = "Failed to register user's U2F key"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category AuthenticationError.",
				reason))
		}

	case ServerError:
		switch reason {
		case QRCodeFailure:
			msg = "Error generating QR code image"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category ServerError.",
				reason))
		}

	case TOTPError:
		switch reason {
		case InvalidTOTP:
			msg = "TOTP is invalid"
		case ProblemEncodingQRCode:
			msg = "There was an error while encoding QR Code"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category TOTPError.",
				reason))
		}

	default:
		panic(fmt.Sprintf("Unsupported error type: %d.",
			category))
	}
	return &Error{ErrorCode: errorCode, Message: msg}
}

// The error interface implementation, which formats to a JSON object string.
func (e *Error) Error() string {
	marshaled, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(marshaled)

}

// Wrap returns an error that contains the given error and an error code derived from
// the given category, reason and the error. Currently, to avoid confusion, it is not
// allowed to create an error of category Success
func Wrap(category Category, reason Reason, err error) *Error {
	errorCode := int(category) + int(reason)
	if err == nil {
		panic("Wrap needs a supplied error to initialize.")
	}

	// do not double wrap a error
	switch err.(type) {
	case *Error:
		panic("Unable to wrap a wrapped error.")
	}

	return &Error{ErrorCode: errorCode, Message: err.Error()}
}
