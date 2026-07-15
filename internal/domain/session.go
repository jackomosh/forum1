package domain

import "time"

type SessionID string

type Session struct {
	ID        SessionID
	UserID    UserID
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AuthenticatedUser struct {
	User    PublicUser
	Session Session
}
