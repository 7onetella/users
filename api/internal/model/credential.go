package model

import (
	"github.com/google/uuid"
)

type UserCredential struct {
	ID         string `db:"credential_id"`
	UserID     string `db:"user_id"`
	Credential string `db:"credential"`
}

func NewUserCredential(userID, credential string) UserCredential {
	return UserCredential{
		ID:         uuid.New().String(),
		UserID:     userID,
		Credential: credential,
	}
}
