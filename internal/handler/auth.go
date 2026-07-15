package handler

import "forum/internal/domain"

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthViewData struct {
	CurrentUser domain.PublicUser
	Error       string
	Form        RegisterRequest
}
