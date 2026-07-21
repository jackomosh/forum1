package handler

import "forum/internal/domain"

type CommentForm struct {
	PostID domain.PostID
	Body   string
}

type CommentViewData struct {
	BaseViewData
	Comment domain.CommentWithAuthor
}
