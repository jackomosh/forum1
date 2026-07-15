package handler

import "forum/internal/domain"

type CommentForm struct {
	PostID domain.PostID
	Body   string
}

type CommentViewData struct {
	CurrentUser domain.PublicUser
	Comment     domain.CommentWithAuthor
	Error       string
}
