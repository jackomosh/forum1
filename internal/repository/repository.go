package repository

import "forum/internal/domain"

type UserRecord struct {
	User domain.User
}

type SessionRecord struct {
	Session domain.Session
	User    domain.User
}

type PostRecord struct {
	Post       domain.Post
	Author     domain.User
	Categories []domain.Category
	Stats      domain.PostStats
	UserVote   domain.VoteValue
}

type CommentRecord struct {
	Comment  domain.Comment
	Author   domain.User
	Stats    domain.CommentStats
	UserVote domain.VoteValue
}

type PostQuery struct {
	Filter domain.PostFilter
}

type CommentQuery struct {
	PostID domain.PostID
	Limit  int
	Offset int
}

type VoteQuery struct {
	UserID   domain.UserID
	Target   domain.VoteTarget
	TargetID int64
}
