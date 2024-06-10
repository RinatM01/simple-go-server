package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Post struct {
	ID   int
	Body string
}

var (
	posts     = make(map[int]Post)
	nextId    = 1
	postsLock sync.Mutex
)

const PORT = ":8080"

func main() {
	http.HandleFunc("/posts", handlerPosts)
	http.HandleFunc("/posts/", handlerPost)

	log.Println("Started on port", PORT)
	fmt.Println("To close connection CTRL+C :-)")

	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handlerPosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetPosts(w, r)
	case "POST":
		handlePostPosts(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePostPosts(w http.ResponseWriter, r *http.Request) {
	var p Post

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading req body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	postsLock.Lock()
	defer postsLock.Unlock()

	currId := nextId
	nextId++
	posts[currId] = p

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	// Locks our local db to prevent data races
	postsLock.Lock()
	defer postsLock.Unlock()

	ps := make([]Post, 0, len(posts))
	for _, v := range posts {
		ps = append(ps, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/posts/"):])
	if err != nil {
		http.Error(w, "Bad post ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleGetPost(w, r, id)
	case "DELETE":
		handleDeletePost(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDeletePost(w http.ResponseWriter, r *http.Request, id int) {
	postsLock.Lock()
	defer postsLock.Unlock()

	if _, ok := posts[id]; !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	delete(posts, id)
	w.WriteHeader(http.StatusOK)
}

func handleGetPost(w http.ResponseWriter, r *http.Request, id int) {
	postsLock.Lock()
	defer postsLock.Unlock()

	p, prs := posts[id]
	if !prs {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
