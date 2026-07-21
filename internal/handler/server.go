package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

type Options struct {
	SessionCookieName string
	SessionDuration   time.Duration
	SessionSecure     bool
	SessionSameSite   string
	PasswordMinLength int
}

type ForumHandler struct {
	repos             repository.Repository
	renderer          *Renderer
	sessionCookieName string
	sessionDuration   time.Duration
	sessionSecure     bool
	sessionSameSite   http.SameSite
	passwordMinLength int
}

type StaticViewData struct {
	BaseViewData
	Categories   []domain.Category
	FeaturedPost *domain.PostWithAuthor
	LatestPosts  []domain.PostWithAuthor
}

func NewForumHandler(repos repository.Repository, renderer *Renderer, opts Options) *ForumHandler {
	sessionDuration := opts.SessionDuration
	if sessionDuration == 0 {
		sessionDuration = 7 * 24 * time.Hour
	}

	cookieName := opts.SessionCookieName
	if cookieName == "" {
		cookieName = "forum_session"
	}

	minPasswordLength := opts.PasswordMinLength
	if minPasswordLength == 0 {
		minPasswordLength = 8
	}

	return &ForumHandler{
		repos:             repos,
		renderer:          renderer,
		sessionCookieName: cookieName,
		sessionDuration:   sessionDuration,
		sessionSecure:     opts.SessionSecure,
		sessionSameSite:   parseSameSite(opts.SessionSameSite),
		passwordMinLength: minPasswordLength,
	}
}

func (h *ForumHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/about", h.About)
	mux.HandleFunc("/dashboard", h.Dashboard)
	mux.HandleFunc("/posts", h.PostsRedirect)
	mux.HandleFunc("/post/create", h.CreatePost)
	mux.HandleFunc("/post/comment", h.CreateComment)
	mux.HandleFunc("/contact", h.Contact)
	mux.HandleFunc("/subscribe", h.Subscribe)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/forgot-password", h.ForgotPassword)
	// Likes and heartbreak votes
	mux.HandleFunc("/post/vote", h.VotePost)
	mux.HandleFunc("/post/", h.PostDetail)
	mux.HandleFunc("/comment/vote", h.CommentVote)
}

func (h *ForumHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	ctx := r.Context()
	categories, err := h.repos.Categories().GetAll(ctx)
	if err != nil {
		h.serverError(w, err)
		return
	}

	filter := domain.PostFilter{
		Kind:      domain.PostFilterAll,
		Sort:      domain.SortNewest,
		Timeframe: domain.TimeframeAll,
		Limit:     4,
		Offset:    0,
	}
	posts, _, err := h.repos.Posts().List(ctx, filter)
	if err != nil {
		log.Printf("❌ Home list error: %v", err)
		h.serverError(w, err)
		return
	}

	var featured *domain.PostWithAuthor
	var latest []domain.PostWithAuthor
	if len(posts) > 0 {
		featured = &posts[0]
		if len(posts) > 1 {
			latest = posts[1:]
		}
	}

	h.renderer.Render(w, "index.html", StaticViewData{
		BaseViewData: BaseViewData{CurrentUser: user},
		Categories:   categories,
		FeaturedPost: featured,
		LatestPosts:  latest,
	})
}

func (h *ForumHandler) About(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	h.renderer.Render(w, "about.html", StaticViewData{
		BaseViewData: BaseViewData{CurrentUser: user},
	})
}

func (h *ForumHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	categories, err := h.repos.Categories().GetAll(ctx)
	if err != nil {
		h.serverError(w, err)
		return
	}

	activeCategory := strings.TrimSpace(r.URL.Query().Get("category"))
	activeFilter := strings.TrimSpace(r.URL.Query().Get("filter"))

	filter := domain.PostFilter{
		Kind:      domain.PostFilterAll,
		Sort:      domain.SortNewest,
		Timeframe: domain.TimeframeAll,
		Limit:     20,
		Offset:    0,
	}
	if user != nil {
		filter.ViewerID = user.ID
	}

	if activeCategory != "" {
		categoryID, ok := categoryIDBySlug(categories, activeCategory)
		if ok {
			filter.Kind = domain.PostFilterCategory
			filter.CategoryID = categoryID
		}
	}

	switch activeFilter {
	case string(domain.PostFilterCreated):
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		filter.Kind = domain.PostFilterCreated
	case string(domain.PostFilterLiked):
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		filter.Kind = domain.PostFilterLiked
	default:
		activeFilter = ""
	}

	// Parse timeframe
	activeTimeframe := strings.TrimSpace(r.URL.Query().Get("timeframe"))
	switch activeTimeframe {
	case "daily":
		filter.Timeframe = domain.TimeframeDaily
	case "weekly":
		filter.Timeframe = domain.TimeframeWeekly
	case "monthly":
		filter.Timeframe = domain.TimeframeMonthly
	default:
		filter.Timeframe = domain.TimeframeAll
	}

	posts, total, err := h.repos.Posts().List(ctx, filter)
	if err != nil {
		log.Printf("❌ List error: %v", err)
		h.serverError(w, err)
		return
	}
	log.Printf("📊 List returned %d posts (total %d)", len(posts), total)

	items := make([]PostListItem, 0, len(posts))
	for _, post := range posts {
		comments, _, err := h.repos.Comments().GetByPostID(ctx, post.Post.ID, 20, 0)
		if err != nil {
			h.serverError(w, err)
			return
		}

		items = append(items, PostListItem{
			Post:     post,
			Comments: comments,
		})

		// Pass user's vote status for each comment if logged in
		if user != nil {
			for i := range comments {
				vote, err := h.repos.Votes().GetVote(ctx, user.ID, domain.VoteTargetComment, int64(comments[i].Comment.ID))
				if err == nil && vote != nil {
					comments[i].UserVote = vote.Value
				}
			}
		}
	}

	h.renderer.Render(w, "dashboard.html", PostListViewData{
		BaseViewData:    BaseViewData{CurrentUser: user},
		Posts:           items,
		Categories:      categories,
		Filter:          filter,
		ActiveCat:       activeCategory,
		ActiveFilter:    activeFilter,
		ActiveTimeframe: activeTimeframe,
	})
}

func (h *ForumHandler) PostsRedirect(w http.ResponseWriter, r *http.Request) {
	target := "/dashboard"
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (h *ForumHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		user, ok := h.requireCurrentUser(w, r)
		if !ok {
			return
		}

		categories, err := h.repos.Categories().GetAll(r.Context())
		if err != nil {
			h.serverError(w, err)
			return
		}

		h.renderer.Render(w, "post_create.html", CreatePostViewData{
			BaseViewData: BaseViewData{CurrentUser: user},
			Categories:   categories,
		})
	case http.MethodPost:
		user, ok := h.requireCurrentUser(w, r)
		if !ok {
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		form, err := h.parsePostForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		draft := domain.PostDraft{
			AuthorID:    user.ID,
			Title:       form.Title,
			Body:        form.Body,
			CategoryIDs: form.CategoryIDs,
		}

		post := &domain.Post{
			AuthorID: draft.AuthorID,
			Title:    draft.Title,
			Body:     draft.Body,
			Status:   domain.PostStatusPublished,
		}

		if err := h.repos.Posts().Create(r.Context(), post, draft.CategoryIDs); err != nil {
			h.serverError(w, err)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ForumHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	postID, err := parseInt64FormValue[domain.PostID](r, "post_id")
	if err != nil {
		http.Error(w, "invalid post", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.FormValue("comment_body"))
	if body == "" {
		http.Error(w, "comment body is required", http.StatusBadRequest)
		return
	}

	draft := domain.CommentDraft{
		PostID:   postID,
		AuthorID: user.ID,
		Body:     body,
	}

	comment := &domain.Comment{
		PostID:   draft.PostID,
		AuthorID: draft.AuthorID,
		Body:     draft.Body,
		Status:   domain.CommentStatusVisible,
	}

	if err := h.repos.Comments().Create(r.Context(), comment); err != nil {
		h.serverError(w, err)
		return
	}

	// Redirect to the specified page, defaulting to dashboard
	redirectTo := r.FormValue("redirect_to")
	if redirectTo == "" {
		redirectTo = "/dashboard"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (h *ForumHandler) Contact(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		log.Printf("contact message from %s %s <%s>",
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("email"),
		)
		http.Redirect(w, r, "/contact?success=true", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.renderer.Render(w, "contact.html", StaticViewData{
		BaseViewData: BaseViewData{CurrentUser: user},
	})
}

func (h *ForumHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("newsletter subscription for %s", r.FormValue("email"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *ForumHandler) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.renderAuth(w, r, "login.html", "")
	case http.MethodPost:
		h.handleLogin(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ForumHandler) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.renderAuth(w, r, "register.html", "")
	case http.MethodPost:
		h.handleRegister(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ForumHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.sessionCookieName)
	if err == nil {
		_ = h.repos.Sessions().Delete(r.Context(), domain.SessionID(cookie.Value))
	}

	http.SetCookie(w, h.expiredSessionCookie())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *ForumHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.renderAuth(w, r, "forgot-password.html", "Password reset email delivery is not configured yet.")
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.renderAuth(w, r, "forgot-password.html", "")
}

// VotePost handles like/dislike votes on posts.
func (h *ForumHandler) VotePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	postID, err := parseInt64FormValue[domain.PostID](r, "post_id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	voteStr := strings.TrimSpace(r.FormValue("vote_val"))
	rawVal, err := strconv.ParseInt(voteStr, 10, 64)
	if err != nil || (rawVal != 1 && rawVal != -1) {
		http.Error(w, "invalid vote value", http.StatusBadRequest)
		return
	}
	voteVal := domain.VoteValue(rawVal)

	ctx := r.Context()

	existing, err := h.repos.Votes().GetVote(ctx, user.ID, domain.VoteTargetPost, int64(postID))
	if err != nil {
		h.serverError(w, err)
		return
	}

	if existing != nil && existing.Value == voteVal {
		if err := h.repos.Votes().RemoveVote(ctx, user.ID, domain.VoteTargetPost, int64(postID)); err != nil {
			h.serverError(w, err)
			return
		}
	} else {
		vote := &domain.Vote{
			UserID:    user.ID,
			Target:    domain.VoteTargetPost,
			TargetID:  int64(postID),
			Value:     voteVal,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := h.repos.Votes().AddVote(ctx, vote); err != nil {
			h.serverError(w, err)
			return
		}
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/dashboard"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

// PostDetail displays a single post with its comments.
func (h *ForumHandler) PostDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/post/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	postID := domain.PostID(id)

	ctx := r.Context()
	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	post, err := h.repos.Posts().GetByID(ctx, postID)
	if err != nil {
		h.serverError(w, err)
		return
	}
	if post == nil {
		http.NotFound(w, r)
		return
	}

	comments, _, err := h.repos.Comments().GetByPostID(ctx, postID, 0, 0)
	if err != nil {
		h.serverError(w, err)
		return
	}

	if user != nil {
		vote, err := h.repos.Votes().GetVote(ctx, user.ID, domain.VoteTargetPost, int64(postID))
		if err == nil && vote != nil {
			post.UserVote = vote.Value
		}
	}

	// Pass user's vote status for each comment if logged in
	if user != nil {
		for i := range comments {
			vote, err := h.repos.Votes().GetVote(ctx, user.ID, domain.VoteTargetComment, int64(comments[i].Comment.ID))
			if err == nil && vote != nil {
				comments[i].UserVote = vote.Value
			}
		}
	}

	data := PostDetailViewData{
		BaseViewData: BaseViewData{CurrentUser: user},
		Post:         *post,
		Comments:     comments,
	}

	h.renderer.Render(w, "posts_detail.html", data)
}

// CommentVote handles like/dislike votes on comments.
func (h *ForumHandler) CommentVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	commentID, err := parseInt64FormValue[domain.CommentID](r, "comment_id")
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	voteStr := strings.TrimSpace(r.FormValue("vote_val"))
	rawVal, err := strconv.ParseInt(voteStr, 10, 64)
	if err != nil || (rawVal != 1 && rawVal != -1) {
		http.Error(w, "invalid vote value", http.StatusBadRequest)
		return
	}
	voteVal := domain.VoteValue(rawVal)

	ctx := r.Context()

	existing, err := h.repos.Votes().GetVote(ctx, user.ID, domain.VoteTargetComment, int64(commentID))
	if err != nil {
		h.serverError(w, err)
		return
	}

	if existing != nil && existing.Value == voteVal {
		if err := h.repos.Votes().RemoveVote(ctx, user.ID, domain.VoteTargetComment, int64(commentID)); err != nil {
			h.serverError(w, err)
			return
		}
	} else {
		vote := &domain.Vote{
			UserID:    user.ID,
			Target:    domain.VoteTargetComment,
			TargetID:  int64(commentID),
			Value:     voteVal,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := h.repos.Votes().AddVote(ctx, vote); err != nil {
			h.serverError(w, err)
			return
		}
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/dashboard"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func (h *ForumHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	identifier := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	if identifier == "" || password == "" {
		h.renderAuth(w, r, "login.html", "Username/email and password are required.")
		return
	}

	user, err := h.lookupUser(r, identifier)
	if err != nil {
		h.serverError(w, err)
		return
	}
	if user == nil || !verifyPassword(user.PasswordHash, password) {
		h.renderAuth(w, r, "login.html", "Invalid username or password credentials.")
		return
	}

	if err := h.startSession(w, r, user.ID); err != nil {
		h.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *ForumHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req := RegisterRequest{
		Username: strings.TrimSpace(r.FormValue("username")),
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: r.FormValue("password"),
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.renderAuthWithForm(w, r, "register.html", "All registration fields are required.", req)
		return
	}
	if len(req.Password) < h.passwordMinLength {
		h.renderAuthWithForm(w, r, "register.html", fmt.Sprintf("Password must be at least %d characters.", h.passwordMinLength), req)
		return
	}

	exists, err := h.repos.Users().ExistsByUsername(r.Context(), req.Username)
	if err != nil {
		h.serverError(w, err)
		return
	}
	if exists {
		h.renderAuthWithForm(w, r, "register.html", "That username is already taken.", req)
		return
	}

	exists, err = h.repos.Users().ExistsByEmail(r.Context(), req.Email)
	if err != nil {
		h.serverError(w, err)
		return
	}
	if exists {
		h.renderAuthWithForm(w, r, "register.html", "That email is already registered.", req)
		return
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		h.serverError(w, err)
		return
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         domain.UserRoleMember,
	}

	if err := h.repos.Users().Create(r.Context(), user); err != nil {
		h.serverError(w, err)
		return
	}

	if err := h.startSession(w, r, user.ID); err != nil {
		h.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *ForumHandler) renderAuth(w http.ResponseWriter, r *http.Request, tmpl, message string) {
	h.renderAuthWithForm(w, r, tmpl, message, RegisterRequest{})
}

func (h *ForumHandler) renderAuthWithForm(w http.ResponseWriter, r *http.Request, tmpl, message string, form RegisterRequest) {
	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return
	}

	h.renderer.Render(w, tmpl, AuthViewData{
		BaseViewData: BaseViewData{
			CurrentUser: user,
			Error:       message,
		},
		Form: form,
	})
}

func (h *ForumHandler) parsePostForm(r *http.Request) (PostForm, error) {
	title := strings.TrimSpace(r.FormValue("title"))
	rawBody := strings.TrimSpace(r.FormValue("body"))
	if title == "" || rawBody == "" {
		return PostForm{}, fmt.Errorf("title and body are required")
	}

	// Preserve formatted HTML from the frontend editor directly
	sanitizedBody := rawBody

	var categoryIDs []domain.CategoryID
	for _, rawID := range r.Form["categories"] {
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || id <= 0 {
			return PostForm{}, fmt.Errorf("invalid category")
		}
		categoryIDs = append(categoryIDs, domain.CategoryID(id))
	}

	if len(categoryIDs) == 0 {
		categories, err := h.repos.Categories().GetAll(r.Context())
		if err != nil {
			return PostForm{}, err
		}
		if id, ok := categoryIDBySlug(categories, "general"); ok {
			categoryIDs = append(categoryIDs, id)
		}
	}

	return PostForm{
		Title:       title,
		Body:        sanitizedBody,
		CategoryIDs: categoryIDs,
	}, nil
}

func (h *ForumHandler) requireCurrentUser(w http.ResponseWriter, r *http.Request) (*domain.PublicUser, bool) {
	user, err := h.currentUser(r)
	if err != nil {
		h.serverError(w, err)
		return nil, false
	}
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}
	return user, true
}

func (h *ForumHandler) currentUser(r *http.Request) (*domain.PublicUser, error) {
	cookie, err := r.Cookie(h.sessionCookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	session, err := h.repos.Sessions().GetByID(r.Context(), domain.SessionID(cookie.Value))
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	if time.Now().After(session.ExpiresAt) {
		_ = h.repos.Sessions().Delete(r.Context(), session.ID)
		return nil, nil
	}

	user, err := h.repos.Users().GetByID(r.Context(), session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	publicUser := publicUserFromDomain(*user)
	return &publicUser, nil
}

func (h *ForumHandler) lookupUser(r *http.Request, identifier string) (*domain.User, error) {
	if strings.Contains(identifier, "@") {
		return h.repos.Users().GetByEmail(r.Context(), identifier)
	}
	return h.repos.Users().GetByUsername(r.Context(), identifier)
}

func (h *ForumHandler) startSession(w http.ResponseWriter, r *http.Request, userID domain.UserID) error {
	token, err := randomToken(32)
	if err != nil {
		return err
	}

	now := time.Now()
	session := &domain.Session{
		ID:        domain.SessionID(token),
		UserID:    userID,
		ExpiresAt: now.Add(h.sessionDuration),
		CreatedAt: now,
	}

	if err := h.repos.Sessions().Create(r.Context(), session); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCookieName,
		Value:    string(session.ID),
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(h.sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   h.sessionSecure,
		SameSite: h.sessionSameSite,
	})
	return nil
}

func (h *ForumHandler) expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     h.sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.sessionSecure,
		SameSite: h.sessionSameSite,
	}
}

func (h *ForumHandler) serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func randomToken(size int) (string, error) {
	token := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func publicUserFromDomain(user domain.User) domain.PublicUser {
	return domain.PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func categoryIDBySlug(categories []domain.Category, slug string) (domain.CategoryID, bool) {
	for _, category := range categories {
		if category.Slug == slug {
			return category.ID, true
		}
	}
	return 0, false
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

type signedInteger interface {
	~int64
}

func parseInt64FormValue[T signedInteger](r *http.Request, key string) (T, error) {
	raw := strings.TrimSpace(r.FormValue(key))
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return T(value), nil
}