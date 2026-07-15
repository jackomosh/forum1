package domain

import "time"

type CommentID int64

type CommentStatus string

const (
	CommentStatusVisible CommentStatus = "visible"
	CommentStatusHidden  CommentStatus = "hidden"
)

type Comment struct {
	ID        CommentID
	PostID    PostID
	AuthorID  UserID
	Body      string
	Status    CommentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CommentWithAuthor struct {
	Comment  Comment
	Author   PublicUser
	Stats    CommentStats
	UserVote VoteValue
}

type CommentStats struct {
	LikeCount    int
	DislikeCount int
	Score        int
}

type CommentDraft struct {
	PostID   PostID
	AuthorID UserID
	Body     string
}

type VoteTarget string

const (
	VoteTargetPost    VoteTarget = "post"
	VoteTargetComment VoteTarget = "comment"
)

type VoteValue int

const (
	VoteNone    VoteValue = 0
	VoteDislike VoteValue = -1
	VoteLike    VoteValue = 1
)

type Vote struct {
	UserID    UserID
	Target    VoteTarget
	TargetID  int64
	Value     VoteValue
	CreatedAt time.Time
	UpdatedAt time.Time
}
