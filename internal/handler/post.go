package handler

import "forum/internal/domain"

type PostForm struct {
	Title       string
	Body        string
	CategoryIDs []domain.CategoryID
}

type PostListViewData struct {
	CurrentUser domain.PublicUser
	Posts       []domain.PostWithAuthor
	Categories  []domain.Category
	Filter      domain.PostFilter
}

type PostDetailViewData struct {
	CurrentUser domain.PublicUser
	Post        domain.PostWithAuthor
	Comments    []domain.CommentWithAuthor
	CommentForm CommentForm
}
