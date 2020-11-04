package model

import (
	"fmt"
)

type JSONAPIErrors struct {
	Errors []JSONAPIError `json:"errors"`
}

type JSONAPIError struct {
	StatusCode int `json:"status,omitempty" example:"401"`
	//Code       string `json:"code,omitempty"`
	//Title      string `json:"title,omitempty"`
	//Detail     string `json:"detail,omitempty"`
	Meta *Error `json:"meta,omitempty"`
}

type Error struct {
	ErrorCode      int    `json:"code"    example:"1100"`
	ErrorCodeHuman string `json:"reason"  example:"database_query_failed"`
	Message        string `json:"message" example:"Database query failed. Check with your administrator."`
	Err            error  `json:"-"`
}

// Credit goes to cloudflare https://github.com/cloudflare/cfssl Error

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

	JSONError // 3XXX

	AuthenticationError // 4xxx
)

// DatabaseError reasons
const (
	// QueryingFailed indicates general database error during SQL execution
	QueryingFailed Reason = 100 * (iota + 1) //11XX

	Unknown
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
)

func New(category Category, reason Reason) *Error {
	errorCode := int(category) + int(reason)
	var msg string
	var errCodeHuman string

	switch category {
	case DatabaseError:
		switch reason {
		case QueryingFailed:
			msg = "Database query failed"
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

	case JSONError:
		switch reason {
		case Marshalling:
			msg = "Error occurred during marshalling"
		case Unmarshalling:
			msg = "Error occurred during unmarshalling"
			errCodeHuman = "error_unmarshalling"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category JSONError.",
				reason))
		}

	case AuthenticationError:
		switch reason {
		case SigninSessionTokenDecodingFailed:
			msg = "Error decoding signin_session_token"
		case SigninSessionExpired:
			msg = "Your login session timed out"
			errCodeHuman = "login_auth_expired"
		case UsernameOrPasswordDoesNotMatch:
			msg = "Check login name or password"
			errCodeHuman = "login_password_invalid"
		case WebauthnAuthFailure:
			msg = "Webauthn authentication failed"
			errCodeHuman = "invalid_webauthn_session_token"
		case TOTPAuthFailure:
			msg = "Your code is invalid"
			errCodeHuman = "login_totp_invalid"
		case UserUnknown:
			msg = "User is unknown"
			errCodeHuman = "login_user_unknown"
		case JWTEncodingFailure:
			msg = "Problem with encoding jwt token"
		default:
			panic(fmt.Sprintf("Unsupported error reason %d under category AuthenticationError.",
				reason))
		}

	default:
		panic(fmt.Sprintf("Unsupported error type: %d.",
			category))
	}
	return &Error{ErrorCode: errorCode, ErrorCodeHuman: errCodeHuman, Message: msg}
}
