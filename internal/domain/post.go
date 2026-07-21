package domain

import (
	"time"
	"html/template"
)

type (
	PostID     int64
	CategoryID int64
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusHidden    PostStatus = "hidden"
)

type Category struct {
	ID          CategoryID
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
}

type Post struct {
	ID        PostID
	AuthorID  UserID
	Title     string
	Body      string
	Status    PostStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostWithAuthor struct {
	Post       Post
	Author     PublicUser
	Categories []Category
	Stats      PostStats
	UserVote   VoteValue
}

type PostStats struct {
	CommentCount int
	LikeCount    int
	DislikeCount int
	Score        int
}

type PostDraft struct {
	AuthorID    UserID
	Title       string
	Body        string
	CategoryIDs []CategoryID
}

type PostFilter struct {
	AuthorID   UserID
	CategoryID CategoryID
	ViewerID   UserID
	Kind       PostFilterKind
	Search     string
	Sort       SortOrder
	Timeframe  Timeframe
	Limit      int
	Offset     int
}

type PostFilterKind string

const (
	PostFilterAll      PostFilterKind = "all"
	PostFilterCategory PostFilterKind = "category"
	PostFilterCreated  PostFilterKind = "created"
	PostFilterLiked    PostFilterKind = "liked"
)

type SortOrder string

const (
	SortNewest      SortOrder = "newest"
	SortOldest      SortOrder = "oldest"
	SortMostLiked   SortOrder = "most_liked"
	SortMostComment SortOrder = "most_commented"
)

type Timeframe string

const (
	TimeframeAll     Timeframe = "all"
	TimeframeDaily   Timeframe = "daily"
	TimeframeWeekly  Timeframe = "weekly"
	TimeframeMonthly Timeframe = "monthly"
)


func (p Post) FormattedBody() template.HTML {
	return template.HTML(p.Body)
}