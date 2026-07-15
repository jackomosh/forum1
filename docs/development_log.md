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