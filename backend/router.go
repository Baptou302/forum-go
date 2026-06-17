package backend

import (
	"net/http"
	"strings"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/post/", postRouter)
	mux.HandleFunc("/comment/", commentRouter)
	mux.HandleFunc("/new-post", newPostHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func postRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch {
	case len(parts) == 2:
		postViewHandler(w, r)
	case len(parts) == 3 && parts[2] == "comment":
		commentHandler(w, r)
	case len(parts) == 3 && parts[2] == "react":
		reactPostHandler(w, r)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

func commentRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 3 && parts[2] == "react" {
		reactCommentHandler(w, r)
	} else {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}
