package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	// "time"
)

type Category struct {
	ID          int
	Name        string
	Slug        string
	Description string
}

type User struct {
	ID       int
	Username string
	Email    string
	Role     string
}

type Comment struct {
	ID        int
	Author    string
	Body      string
	CreatedAt string
}

type Post struct {
	ID         int
	Title      string
	Body       string
	AuthorName string
	Categories []string
	Likes      int
	Dislikes   int
	UserVoted  int // 1 for liked, -1 for disliked, 0 for none
	Comments   []Comment
	CreatedAt  string
}

// Global page structural model
type PageData struct {
	User         *User
	Categories   []Category
	Posts        []Post
	ActiveCat    string
	ActiveFilter string
	Error        string
}

// --- Shared In-Memory Database State ---
var mockCategories = []Category{
	{ID: 1, Name: "Technology", Slug: "technology", Description: "All about the latest tech"},
	{ID: 2, Name: "Science", Slug: "science", Description: "Scientific discoveries"},
	{ID: 3, Name: "Art", Slug: "art", Description: "Creative arts and design"},
	{ID: 4, Name: "Gaming", Slug: "gaming", Description: "Video games and culture"},
	{ID: 5, Name: "General", Slug: "general", Description: "General discussion topics"},
}

// This slice is shared by both the root "/" index page and the "/dashboard" page
var globalPosts = []Post{
	{
		ID:         1,
		Title:      "Why SQLite is the optimal choice for local-first apps",
		Body:       "SQLite's performance with WAL mode enabled is incredibly fast. For smaller platforms or local development environments like our DevForum, its file-backed architecture bypasses network overhead completely! What is your experience with SQLite performance?",
		AuthorName: "jacomondi",
		Categories: []string{"Technology", "General"},
		Likes:      42,
		Dislikes:   2,
		UserVoted:  1,
		CreatedAt:  "2 hours ago",
		Comments: []Comment{
			{ID: 1, Author: "tech_guru", Body: "WAL mode is definitely a game-changer! Highly recommend it.", CreatedAt: "1 hour ago"},
			{ID: 2, Author: "go_dev", Body: "Simple, self-contained, and matches perfect concurrency rules if configured right.", CreatedAt: "45 mins ago"},
		},
	},
}

var loggedInUser = &User{
	ID:       1,
	Username: "jacomondi",
	Email:    "jacomondi@learn.zone01kisumu.ke",
	Role:     "member",
}

// Standard parsing and execution utility
func render(w http.ResponseWriter, tmpl string, data PageData) {
	files := []string{
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", tmpl),
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("Template compilation error: %v", err)
		http.Error(w, "Failed to compile template files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route: Public Landing Page (Index)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Serves index.html with the dynamically updated globalPosts
		render(w, "index.html", PageData{
			User:       nil, // Simulated public guest
			Categories: mockCategories,
			Posts:      globalPosts,
		})
	})

	// Route: Logged-in Core Forum Dashboard Page
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		selectedCategory := r.URL.Query().Get("category")
		selectedFilter := r.URL.Query().Get("filter")

		// Filter posts by category or user context
		var filteredPosts []Post
		for _, post := range globalPosts {
			matchesCategory := selectedCategory == ""
			if selectedCategory != "" {
				for _, catName := range post.Categories {
					for _, c := range mockCategories {
						if c.Slug == selectedCategory && c.Name == catName {
							matchesCategory = true
							break
						}
					}
				}
			}

			matchesFilter := true
			if selectedFilter == "created" && post.AuthorName != loggedInUser.Username {
				matchesFilter = false
			}
			if selectedFilter == "liked" && post.UserVoted != 1 {
				matchesFilter = false
			}

			if matchesCategory && matchesFilter {
				filteredPosts = append(filteredPosts, post)
			}
		}

		render(w, "dashboard.html", PageData{
			User:         loggedInUser,
			Categories:   mockCategories,
			Posts:        filteredPosts,
			ActiveCat:    selectedCategory,
			ActiveFilter: selectedFilter,
		})
	})

	// Action Route: Form handler for creating posts
	http.HandleFunc("/post/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		body := r.FormValue("body")
		selectedCatIDs := r.Form["categories"]

		// Map structural strings from selected IDs
		var cats []string
		for _, idStr := range selectedCatIDs {
			id, _ := strconv.Atoi(idStr)
			for _, cat := range mockCategories {
				if cat.ID == id {
					cats = append(cats, cat.Name)
				}
			}
		}
		if len(cats) == 0 {
			cats = append(cats, "General")
		}

		// Save the post into our shared global memory space
		newPost := Post{
			ID:         len(globalPosts) + 1,
			Title:      title,
			Body:       body,
			AuthorName: loggedInUser.Username,
			Categories: cats,
			Likes:      0,
			Dislikes:   0,
			UserVoted:  0,
			CreatedAt:  "Just now",
			Comments:   []Comment{},
		}

		// Insert post at the beginning of our list
		globalPosts = append([]Post{newPost}, globalPosts...)

		// Redirect back to dashboard to see immediate update
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	// Action Route: Form handler to comment on posts
	http.HandleFunc("/post/comment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		postID, _ := strconv.Atoi(r.FormValue("post_id"))
		body := r.FormValue("comment_body")

		if body != "" {
			for i, post := range globalPosts {
				if post.ID == postID {
					newComment := Comment{
						ID:        len(post.Comments) + 1,
						Author:    loggedInUser.Username,
						Body:      body,
						CreatedAt: "Just now",
					}
					globalPosts[i].Comments = append(globalPosts[i].Comments, newComment)
					break
				}
			}
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})

	// Add this route handler into your main() function in cmd/forum/main.go:

http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Form submission logic (Redirects to contact page with a success signal)
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		// In a real database scenario, you would save this to an inbox table here.
		log.Printf("Received message from %s %s (%s)", r.FormValue("first_name"), r.FormValue("last_name"), r.FormValue("email"))
		http.Redirect(w, r, "/contact?success=true", http.StatusSeeOther)
		return
	}

	render(w, "contact.html", PageData{
		User:       loggedInUser, // Pass active user so Navbar highlights profile state correctly
		Categories: mockCategories,
		Posts:      globalPosts,
	})
})

	// Routing mock login states
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		errMsg := ""
		if r.URL.Query().Get("err") == "true" {
			errMsg = "Invalid username or password credentials."
		}
		render(w, "login.html", PageData{Error: errMsg})
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		render(w, "register.html", PageData{})
	})
	http.HandleFunc("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		render(w, "forgot-password.html", PageData{})
	})

	log.Println("=== Front-End Mock Dev Server ===")
	log.Println("Server running at: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}