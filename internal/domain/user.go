package domain

import "time"

type UserID int64

type UserRole string

const (
	UserRoleMember UserRole = "member"
	UserRoleAdmin  UserRole = "admin"
)

type User struct {
	ID           UserID
	Username     string
	Email        string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PublicUser struct {
	ID        UserID
	Username  string
	Role      UserRole
	CreatedAt time.Time
}

type UserRegistration struct {
	Username string
	Email    string
	Password string
}

type UserCredentials struct {
	Email    string
	Password string
}
