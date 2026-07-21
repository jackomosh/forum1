package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

type stubRepository struct {
	categories []domain.Category
	session    *domain.Session
	user       *domain.User
}

func (s *stubRepository) Users() repository.UserRepository { return stubUserRepository{user: s.user} }
func (s *stubRepository) Sessions() repository.SessionRepository {
	return stubSessionRepository{session: s.session}
}
func (s *stubRepository) Posts() repository.PostRepository       { return stubPostRepository{} }
func (s *stubRepository) Comments() repository.CommentRepository { return stubCommentRepository{} }
func (s *stubRepository) Categories() repository.CategoryRepository {
	return stubCategoryRepository{categories: s.categories}
}
func (s *stubRepository) Votes() repository.VoteRepository { return stubVoteRepository{} }
func (s *stubRepository) Close() error                     { return nil }

type stubUserRepository struct{ user *domain.User }

func (s stubUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (s stubUserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if s.user != nil && s.user.ID == id {
		return s.user, nil
	}
	return nil, nil
}
func (s stubUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (s stubUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}
func (s stubUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (s stubUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (s stubUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

type stubSessionRepository struct{ session *domain.Session }

func (s stubSessionRepository) Create(ctx context.Context, session *domain.Session) error { return nil }
func (s stubSessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	if s.session != nil && s.session.ID == id {
		return s.session, nil
	}
	return nil, nil
}
func (s stubSessionRepository) GetByUserID(ctx context.Context, userID domain.UserID) ([]domain.Session, error) {
	return nil, nil
}
func (s stubSessionRepository) Delete(ctx context.Context, id domain.SessionID) error { return nil }
func (s stubSessionRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	return nil
}
func (s stubSessionRepository) CleanupExpired(ctx context.Context) error { return nil }

type stubPostRepository struct{}

func (s stubPostRepository) Create(ctx context.Context, post *domain.Post, categoryIDs []domain.CategoryID) error {
	return nil
}
func (s stubPostRepository) GetByID(ctx context.Context, id domain.PostID) (*domain.PostWithAuthor, error) {
	return nil, nil
}
func (s stubPostRepository) Update(ctx context.Context, post *domain.Post) error { return nil }
func (s stubPostRepository) Delete(ctx context.Context, id domain.PostID) error  { return nil }
func (s stubPostRepository) List(ctx context.Context, filter domain.PostFilter) ([]domain.PostWithAuthor, int, error) {
	return nil, 0, nil
}
func (s stubPostRepository) GetCategoriesByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error) {
	return nil, nil
}

type stubCommentRepository struct{}

func (s stubCommentRepository) Create(ctx context.Context, comment *domain.Comment) error { return nil }
func (s stubCommentRepository) GetByID(ctx context.Context, id domain.CommentID) (*domain.CommentWithAuthor, error) {
	return nil, nil
}
func (s stubCommentRepository) GetByPostID(ctx context.Context, postID domain.PostID, limit, offset int) ([]domain.CommentWithAuthor, int, error) {
	return nil, 0, nil
}
func (s stubCommentRepository) Update(ctx context.Context, comment *domain.Comment) error { return nil }
func (s stubCommentRepository) Delete(ctx context.Context, id domain.CommentID) error     { return nil }

type stubCategoryRepository struct{ categories []domain.Category }

func (s stubCategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	return s.categories, nil
}
func (s stubCategoryRepository) GetByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	return nil, nil
}
func (s stubCategoryRepository) GetByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error) {
	return nil, nil
}
func (s stubCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return nil
}

type stubVoteRepository struct{}

func (s stubVoteRepository) AddVote(ctx context.Context, vote *domain.Vote) error { return nil }
func (s stubVoteRepository) GetVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) (*domain.Vote, error) {
	return nil, nil
}
func (s stubVoteRepository) GetPostStats(ctx context.Context, postID domain.PostID) (*domain.PostStats, error) {
	return nil, nil
}
func (s stubVoteRepository) GetCommentStats(ctx context.Context, commentID domain.CommentID) (*domain.CommentStats, error) {
	return nil, nil
}
func (s stubVoteRepository) RemoveVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) error {
	return nil
}
func (s stubVoteRepository) GetUserVotedPostIDs(ctx context.Context, userID domain.UserID) ([]domain.PostID, error) {
	return nil, nil
}

func TestCreatePostRendersFormOnGet(t *testing.T) {
	renderer := NewRenderer(filepath.Join("..", "..", "web", "templates"))
	repo := &stubRepository{
		categories: []domain.Category{{ID: 1, Name: "General", Slug: "general"}},
		session:    &domain.Session{ID: "sess-1", UserID: 1, ExpiresAt: time.Now().Add(time.Hour)},
		user:       &domain.User{ID: 1, Username: "alice", Role: domain.UserRoleMember},
	}
	handler := NewForumHandler(repo, renderer, Options{SessionCookieName: "forum_session"})

	req := httptest.NewRequest(http.MethodGet, "/post/create", nil)
	req.AddCookie(&http.Cookie{Name: "forum_session", Value: "sess-1"})
	rr := httptest.NewRecorder()

	handler.CreatePost(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%q", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Draft a Technical Thread") {
		t.Fatalf("expected create-post form to render, got %q", rr.Body.String())
	}
}
