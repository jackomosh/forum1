package repository

import (
	"context"

	"forum/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID domain.UserID) ([]domain.Session, error)
	Delete(ctx context.Context, id domain.SessionID) error
	DeleteByUserID(ctx context.Context, userID domain.UserID) error
	CleanupExpired(ctx context.Context) error
}

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post, categoryIDs []domain.CategoryID) error
	GetByID(ctx context.Context, id domain.PostID) (*domain.PostWithAuthor, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id domain.PostID) error
	List(ctx context.Context, filter domain.PostFilter) ([]domain.PostWithAuthor, int, error)
	GetCategoriesByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id domain.CommentID) (*domain.CommentWithAuthor, error)
	GetByPostID(ctx context.Context, postID domain.PostID, limit, offset int) ([]domain.CommentWithAuthor, int, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, id domain.CommentID) error
}

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error)
	GetByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error)
	Create(ctx context.Context, category *domain.Category) error
}

type VoteRepository interface {
	AddVote(ctx context.Context, vote *domain.Vote) error
	GetVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) (*domain.Vote, error)
	GetPostStats(ctx context.Context, postID domain.PostID) (*domain.PostStats, error)
	GetCommentStats(ctx context.Context, commentID domain.CommentID) (*domain.CommentStats, error)
	RemoveVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) error
	GetUserVotedPostIDs(ctx context.Context, userID domain.UserID) ([]domain.PostID, error)
}

// Repository aggregates all repository interfaces
type Repository interface {
	Users() UserRepository
	Sessions() SessionRepository
	Posts() PostRepository
	Comments() CommentRepository
	Categories() CategoryRepository
	Votes() VoteRepository
	Close() error
}
