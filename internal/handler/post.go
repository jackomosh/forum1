package handler

import "forum/internal/domain"

type PostForm struct {
	Title       string
	Body        string
	CategoryIDs []domain.CategoryID
}

type BaseViewData struct {
	CurrentUser *domain.PublicUser
	Error       string
}

type PostListItem struct {
	Post     domain.PostWithAuthor
	Comments []domain.CommentWithAuthor
}

type PostListViewData struct {
	BaseViewData
	Posts        []PostListItem
	Categories   []domain.Category
	Filter       domain.PostFilter
	ActiveCat    string
	ActiveFilter string
	ActiveTimeframe string
}

type CreatePostViewData struct {
	BaseViewData
	Categories []domain.Category
}

type PostDetailViewData struct {
	BaseViewData
	Post        domain.PostWithAuthor
	Comments    []domain.CommentWithAuthor
	CommentForm CommentForm
}
