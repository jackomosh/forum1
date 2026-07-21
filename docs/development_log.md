# Development Log

This log records the main development decisions, implementation steps, and verification results for the forum project. Each contributor should add a new entry when they introduce a feature, change the architecture, adjust data models, or fix an important bug.

## Entry Format

Use this structure for every new entry:

```md
## Day N - Short Title

**Date:** YYYY-MM-DD
**Author:** Name
**Branch:** branch-name

### Goal
Briefly explain what the work was meant to achieve.

### Implementation
- Describe the files or packages changed.
- Explain the important design decisions.
- Mention any constraints or assumptions.

### Verification
- List commands run, manual checks done, or tests added.
- Note anything that could not be tested.

### Next Steps
- List follow-up work that should happen after this entry.
```

Keep entries short but useful. The goal is not to write a diary; the goal is to help the next developer understand what changed, why it changed, and how to continue without guessing.

## Day 1 - Go Data Definitions

**Date:** 2026-07-10

**Author:** [Bramwel Mutugi](https://learn.zone01kisumu.ke/git/mumutugi)

**Branch:** `feature/domain-structs`

### Goal

Set up the first Go layer of the project using pure data definitions only. This branch intentionally avoids application behavior, database queries, HTTP handlers, and business logic so the rest of the team can build on stable shared types.

### Implementation

- Added a minimal `go.mod` with module name `forum` so all packages can import each other consistently.
- Defined configuration structs in `internal/config/config.go`, including application, server, database, session, and security settings.
- Defined core domain models in `internal/domain`, including users, sessions, posts, categories, comments, votes, stats, filters, and draft/input models.
- Defined handler-facing request and view data structs in `internal/handler`, such as auth forms, post views, comment forms, vote requests, pagination data, flash messages, and request context.
- Defined repository-facing data containers in `internal/repository`, keeping them separate from HTTP concerns.
- Defined SQLite row structs in `internal/repository/sqlite` for users, sessions, posts, categories, comments, votes, migrations, and aggregate stats.
- Added small application/dependency structs in `internal/app/app.go` to reserve the future composition shape without mixing definitions into the command entrypoint.

After reviewing the project instructions, the data models were refined to support the required forum behavior:

- Added `PostFilterKind` in `internal/domain/post.go` with `all`, `category`, `created`, and `liked` filter modes.
- Added `ViewerID` to `PostFilter` so created-post and liked-post filters can be scoped to the logged-in user.
- Added `VoteNone` to represent users who have not liked or disliked a post or comment.
- Added `UserVote` to post/comment view models and repository records so public pages can show total likes/dislikes while logged-in users can also see their own vote state.
- Added `UserVote` fields to SQLite stats rows for posts and comments so repository queries can return aggregate counts and the current user's vote together.

The main design decision was to separate data definitions by project layer:

- `internal/domain` holds business entities and shared domain types.
- `internal/config` holds runtime configuration shapes.
- `internal/handler` holds HTTP request and view payload shapes.
- `internal/repository` holds persistence-facing records and query shapes.
- `internal/repository/sqlite` holds database row shapes specific to SQLite.

This keeps the project ready for implementation while avoiding early coupling between handlers, repositories, and domain logic.

### Verification

Formatted the Go code with `gofmt`.

Compiled all packages with:

```sh
env GOCACHE=/tmp/go-build-cache go test -v ./...
```

The command passed across all current packages. There are no test files yet because this entry only introduces data structures.

### Next Steps

- Add the database schema that matches the SQLite row definitions.
- Implement repository interfaces and SQLite queries.
- Implement HTTP handlers using the handler request/view structs.
- Add validation rules for user registration, login, posts, comments, and votes.
- Add tests once behavior is introduced.

## Day 2 - Repository Layer & SQLite Implementation (Complete)

**Date:** 2026-07-13
**Author:** [Stanley Thuita](https://learn.zone01kisumu.ke/git/stathuita)
**Branch:** `feature/sqlite-repository`

### Goal
Implement the complete repository layer with SQLite as the database backend, providing data persistence for all domain models.

### Implementation

#### Database Schema (schema.sql)
Created complete SQLite schema with all tables: users, sessions, posts, categories, post_categories, comments, votes. Added indexes for performance and default categories.

#### Repository Interfaces (internal/repository/interfaces.go)
Defined contracts for all repositories: User, Session, Post, Comment, Category, Vote.

#### SQLite Implementation (internal/repository/sqlite/)
- **client.go**: Connection pooling with WAL mode
- **user.go**: CRUD operations with email/username existence checks
- **session.go**: Session management with expiration cleanup
- **post.go**: Post CRUD with filtering, sorting, and pagination
- **comment.go**: Comment CRUD with post association
- **vote.go**: Vote management with UPSERT pattern and stats aggregation
- **category.go**: Category management

#### Key Design Decisions
- **Timestamps**: Store as Unix timestamps (int64) in database, use time.Time in domain models
- **Foreign Keys**: ON DELETE CASCADE for referential integrity
- **Votes**: UPSERT pattern (INSERT OR REPLACE) for idempotent operations
- **Testing**: In-memory database (:memory:) with embedded schema for isolation

### Verification

```sh
go build ./internal/repository/sqlite
go test -v ./internal/repository/sqlite
```
**Results:**

- All 4 tests passing
- Build successful with no errors
- No file I/O in tests (embedded schema)

**Next Steps**

- Implement bcrypt password hashing (Member 2)
- Create HTTP handlers using repository interfaces (Member 3)
- Implement cookie-based session management (Member 3)
- Complete UI templates (Member 4)
- Finalize Docker containerization (Member 5)

## Day 2 - Dynamic UI Templates & Semantics (I don't know which day it is 😬)

### **Date:** 2026-07-13
### **Author:** [Jack Omondi](https://learn.zone01kisumu.ke/git/jacomondi)
### **Branch:** `feature/ui-templates`

### Goal
Implement a robust, semantic, and reusable Go template engine structure (`web/templates/`) to serve as the unified presentation layer. This layout provides clear dynamic data slots (`{{block}}` / `{{define}}`) for Member 3's handlers and provides predictable class hooks for Member 5's styling.

### Implementation
- **Layout Structuring (`web/templates/base.html`)**: Created the master layout page establishing the HTML5 boilerplate, linking static assets, and defining standard dynamic entry blocks (`{{block "content" .}}`).
- **Dashboard Development (`web/templates/index.html`)**: Implemented the main discussion feed displaying dynamic posts, categories, and post-filtering components while checking user authentication context globally.
- **Authentication Forms (`web/templates/login.html` & `web/templates/register.html`)**: Built native forms for user authentication featuring semantic, accessible `<input>` types matching domain specs.
- **Data Integration Assumptions**: Ensured all dynamic template hooks (e.g., `.User`, `.Posts`, `.Categories`) directly match the view model definitions introduced on Day 1, allowing seamless synchronization with HTTP controllers.

## Front-End Testing & Preview Guide
Allow me to give you an overview of the UI

To easily preview changes, test frontend, and review style updates without requiring a database, router, or complex backend logic, a lightweight mock development server has been set up at `cmd/forum/main.go`. 

#### 1. Running the Mock Server
From your terminal (after cloning the repository [Forum](https://learn.zone01kisumu.ke/git/stathuita/forum) or git pull to get recent updates), navigate to the **project root directory** and run:

```bash
go run ./cmd/forum
```

#### 1. Opening the server
You should see these messages from your terminal if everything goes as expected
- **Date: Time === Front-End Mock Dev Server ===**
- **Date: Time " Server running at: http://localhost:8080 " make sure to navigate to: [localhost](http://localhost:8080)**

### Verification (still-working)
- Validate template syntax using Go template parsing logic.
- Checke HTML markup semantics and structure via raw local file renders.
- Confirm CSS selectors are cleanly structured, enabling Member 5 to begin immediate styling without structural blocks.

### Next Steps (still-working)
- Hand off dynamic templates to Member 3 for integration with Go HTTP route controllers.
- Assist Member 5 with mapping specific DOM class hooks in `style.css` and `main.js`.


## Day 3 - Backend Entrypoint Refactor & Repository Wiring

**Date:** 2026-07-14

**Author:** [Bramwel Mutugi](https://learn.zone01kisumu.ke/git/mumutugi)

**Branch:** `main`

### Goal
Restore `cmd/forum/main.go` to a minimal application entrypoint and move backend behavior into the predefined `internal/` modules, using the existing domain, handler, repository, and SQLite structures instead of redeclared mock structs.

### Implementation
- Reduced `cmd/forum/main.go` to only call application startup through `internal/app`.
- Moved application composition into `internal/app/app.go`, including config loading, SQLite client creation, schema application, repository construction, static file routing, handler registration, and server startup.
- Kept runtime defaults inside the existing `internal/config/config.go` file to preserve the original package structure.
- Added real HTTP handler wiring in `internal/handler/server.go`, with template rendering in `internal/handler/renderer.go` and password helper logic in `internal/handler/password.go`.
- Updated handler view data in `internal/handler/post.go`, `internal/handler/auth.go`, and `internal/handler/comment.go` so templates receive domain-backed data rather than mock structs.
- Added missing SQLite implementations for predefined repository contracts:
  - `internal/repository/sqlite/category.go` implements category persistence.
  - `internal/repository/sqlite/session.go` implements session persistence.
  - `internal/repository/sqlite/store.go` aggregates SQLite repositories behind the existing `repository.Repository` interface.
- Updated existing SQLite repositories to include compile-time interface checks.
- Fixed `PostRepository.List` so the SQLite implementation matches the existing interface and returns user vote state through `domain.VoteValue`.
- Updated `web/templates/base.html` and `web/templates/dashboard.html` to use `.CurrentUser`, `domain.PostWithAuthor`, `domain.CommentWithAuthor`, and `domain.Category` data.
- Adjusted SQLite client settings in `internal/repository/sqlite/client.go` with `_busy_timeout=5000`, foreign-key DSN options, and a single open connection to reduce local SQLite lock contention.

### Verification
- Formatted changed Go files with `gofmt`.
- Ran the full test suite:

```sh
env GOCACHE=/tmp/go-build-cache GOMODCACHE=/tmp/go-mod-cache go test -v ./...
```

- Smoke-tested the server on an alternate port with `FORUM_PORT=18080` because port `8080` was already occupied.
- Confirmed `/`, `/dashboard`, `/login`, and `/register` returned `200`.
- Confirmed protected post/comment actions redirect unauthenticated users to `/login`.

### Next Steps
- Add HTTP routes and forms or API handlers for post/comment voting.
- Add CSRF protection for all state-changing forms.
- Strengthen validation for registration, login, posts, comments, and categories.
- Add repository tests for posts, comments, categories, sessions, and votes.
- Add graceful shutdown for the HTTP server.
- Replace raw schema application on startup with versioned migrations.
- Cache parsed templates instead of reparsing templates on each request.
- Ensure developers close any external `sqlite3 forum.db` shell before running the app to avoid SQLite locks.


## Day 3 - Dynamic UI Templates & Increased Additional Pages

### **Date:** 2026-07-16
### **Author:** [Jack Omondi](https://learn.zone01kisumu.ke/git/jacomondi)
### **Branch:** `feature/ui-templates`

### Goal
Implement additional frontend pages, a robust, semantic, and reusable Go template structure inside (`web/templates/`) to serve as the unified presentation layer.

### Implementation

- **Frontend Pages (`web/templates/posts.html` & `web/templates/post_create.html` & `web/templates/posts_detail.html` )**: Implemented the fronend pages discussion feed displaying dynamic posts, categories, users create posts and edit posts dynamically, while also having the ability to view only one posts discussion once opened.

### Verification (still-working)
- Validate all templates syntax using Go template parsing logic.
- Checke HTML markup semantics and structure via raw local file renders.
- Confirm CSS selectors are cleanly structured, and each page is styled according to use cases and overal behaviour

### Next Steps (still-working)
- Hand off dynamic templates to all members for integration with Go HTTP route controllers or any other backend logics.
- Make sure `style.css` and `main.js` are styling and ensuring responsiveness across pages.

## Day 4 - Menu Buttons

### **Date:** 2026-07-17
### **Author:** [Jack Omondi](https://learn.zone01kisumu.ke/git/jacomondi)
### **Branch:** `feature/ui-templates`

### Goal
Implement additional functioning menu button, across all Go template structure inside (`web/templates/`).

### Implementation

- **Frontend Improvements (`setup an hamburger menu button` )**:

### Verification
- Validated that all Cascading Style Sheet and JavaScript is present for the menu button to work as expected
- Confirmed that these changes indeed make sure the menu is working

### Next Steps
- Fix the post_create page to accept images and render cleanly
- Fix the filter by relevant / trending discussions
- Fix the voting i.e likes and dislikes and connect to the backend
- Insert more categories on the dashboard
- Add a password toggle on the login / register pages
- Fix the Online Developers on the dashboard to be dynamic and not hardcoded
- Fix the posts_detail page to show up now its 404

## Day 5 - Fix Create Post Route Handling

**Date:** 2026-07-18
**Author:** [Bramwel Mutugi](https://learn.zone01kisumu.ke/git/mumutugi)
**Branch:** fix/method_error

### Goal
Resolve the create-post flow from the dashboard, which was returning a 405 Method Not Allowed because the handler only supported POST while the UI routed to the endpoint with GET.

### Implementation
- Updated [internal/handler/server.go](internal/handler/server.go) so the create-post handler now renders the creation form for GET requests and preserves the existing publish behavior for POST requests.
- Added a dedicated view model in [internal/handler/post.go](internal/handler/post.go) for the create-post page.
- Added a regression test in [internal/handler/server_test.go](internal/handler/server_test.go) to ensure the endpoint renders the form successfully on GET.
- Added forum.db to git ignore for good practice.

### Verification
- Ran `go test ./internal/handler`
- Ran `go test ./...`

Both commands completed successfully, and the new regression test passed.

### Next Steps
- Add CSRF protection for state-changing forms.
- Improve validation and error feedback for post creation.
- Continue polishing the dashboard and post creation experience.
