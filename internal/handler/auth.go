package handler

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
	BaseViewData
	Form RegisterRequest
}
