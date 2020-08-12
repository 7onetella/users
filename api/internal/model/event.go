package model

import (
	"github.com/google/uuid"
	"time"
)

type AuthEvent struct {
	ID        string `db:"event_id"`
	UserIDReq string `db:"user_id"`
	Event     string `db:"event"`
	Timestamp int64  `db:"event_timestamp"`
	IPV4      string `db:"ip_v4"`
	IPV6      string `db:"ip_v6"`
	Agent     string `db:"agent"`
}

func NewAuthEvent(userID, event, ipv4, ipv6, agent string) AuthEvent {
	return AuthEvent{
		ID:        uuid.New().String(),
		UserIDReq: userID,
		Event:     event,
		IPV4:      ipv4,
		IPV6:      ipv6,
		Agent:     agent,
		Timestamp: time.Now().Unix(),
	}
}
