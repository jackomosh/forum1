package sqlite

import "forum/internal/repository"

var _ repository.Repository = (*Repository)(nil)

type Repository struct {
	client     *Client
	users      *UserRepository
	sessions   *SessionRepository
	posts      *PostRepository
	comments   *CommentRepository
	categories *CategoryRepository
	votes      *VoteRepository
}

func NewRepository(client *Client) *Repository {
	return &Repository{
		client:     client,
		users:      NewUserRepository(client),
		sessions:   NewSessionRepository(client),
		posts:      NewPostRepository(client),
		comments:   NewCommentRepository(client),
		categories: NewCategoryRepository(client),
		votes:      NewVoteRepository(client),
	}
}

func (r *Repository) Users() repository.UserRepository {
	return r.users
}

func (r *Repository) Sessions() repository.SessionRepository {
	return r.sessions
}

func (r *Repository) Posts() repository.PostRepository {
	return r.posts
}

func (r *Repository) Comments() repository.CommentRepository {
	return r.comments
}

func (r *Repository) Categories() repository.CategoryRepository {
	return r.categories
}

func (r *Repository) Votes() repository.VoteRepository {
	return r.votes
}

func (r *Repository) Close() error {
	return r.client.Close()
}
