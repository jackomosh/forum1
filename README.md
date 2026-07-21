# Forum
A lightweight, feature-rich web forum application built in Go using native HTML templates, SQLite, and Docker. This project enables registered users to communicate via posts and comments, categorize topics, react via likes/dislikes, and filter content, all wrapped inside a containerized environment.

## Features

### User Authentication & Authorization
- Registration & Login:
Secure registration validating email uniqueness, username, and password.

- Session Management:
Single active session handling per user using HTTP-only cookies with configurable expiration times and unique UUIDs.

- Password Encryption:
Secure password hashing using bcrypt before storing credentials in the database.

### Posts & Comments
- Creation:
Registered users can author posts and comment on existing discussions.

- Categorization:
Associate single or multiple categories with a post upon creation.

- Public Visibility:
Posts, comments, categories, and reaction counters are visible to everyone (including non-registered guests).

### Reactions & Filtering
- Likes & Dislikes:
Registered users can like or dislike both posts and comments in real-time.

- Subforum Category Filtering:
Browse posts filtered by specific categories.

- User Filters:
Registered users can filter the feed to show Created Posts or Liked Posts.

### Containerization & Error Handling
- Fully containerized with Docker for easy deployment and consistent execution across environments.

- Robust HTTP status code handling (400 Bad Request, 401 Unauthorized, 404 Not Found, 500 Internal Server Error).

## Tech Stack & Dependencies
Backend: Go (net/http, HTML templates, standard library)

Database: SQLite (go-sqlite3)

Security: golang.org/x/crypto/bcrypt (Password hashing) & google/uuid (Session tokens)

Frontend: Vanilla HTML5, CSS3, and JavaScript (No external JS frameworks used)

Containerization: Docker & Docker Compose

## Database Schema Design
The SQLite database uses an Entity-Relationship model structured as follows:
```text
+---------------+        +---------------+        +---------------+
|     Users     |        |     Posts     |        |  Categories   |
+---------------+        +---------------+        +---------------+
| id (PK)       |<------1| id (PK)       |   +--->| id (PK)       |
| email         |        | user_id (FK)  |   |    | name          |
| username      |        | title         |   |    +---------------+
| password_hash |        | content       |   |
| created_at    |        | created_at    |   |    +-------------------+
+---------------+        +---------------+   |    |  Post_Categories  |
        |                        ^           |    +-------------------+
        |                        |           +---| post_id (FK)      |
        v                        |                | category_id (FK)  |
+---------------+                |                +-------------------+
|   Sessions    |                |
+---------------+                |                +---------------+
| id (PK)       |                |                |   Comments    |
| user_id (FK)  |                |                +---------------+
| session_token |                +---------------1| id (PK)       |
| expires_at    |                                 | post_id (FK)  |
+---------------+                                 | user_id (FK)  |
                                                  | content       |
                                                  +---------------+
```
## Getting Started
Prerequisites
Go (v1.20+) (for local execution)

Docker Engine (recommended for containerized run)

1. Clone the Repository
```Bash
git clone https://github.com/your-username/forum.git
cd forum
```
2. Run via Docker (Recommended)
Build and run the Docker container:

```Bash
# Build the image
docker build -t forum-app .

# Run the container
docker run -p 8080:8080 forum-app
```
Navigate to http://localhost:8080 in your web browser.

3. Run Locally Without Docker
Ensure you have SQLite build tools installed (gcc/CGO_ENABLED=1).

```Bash
# Install dependencies
go mod download

# Run the application
go run main.go
```
The application will start on http://localhost:8080.

## Testing
To run unit tests for the core HTTP handlers and database methods:

```Bash
go test -v ./...
```

## Project Structure
```text
├── cmd/
│   └── main.go          # Application entry point
├── internal/
│   ├── database/        # SQLite connection, migrations, and query execution
│   ├── handlers/        # HTTP handlers for auth, posts, comments, filters
│   └── models/          # Data structs (User, Post, Comment, Reaction)
├── static/              # CSS styles and static client assets
├── templates/           # HTML templates
├── Dockerfile           # Docker container build instructions
├── go.mod               # Go module definition
└── README.md
```
## License
This project was developed for educational purposes as part of the software engineering curriculum.