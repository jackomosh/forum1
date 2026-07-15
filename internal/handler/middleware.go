package handler

import "forum/internal/domain"

type RequestContext struct {
	RequestID string
	User      domain.PublicUser
	SessionID domain.SessionID
	CSRFToken string
}

type FlashMessage struct {
	Kind    string
	Message string
}
