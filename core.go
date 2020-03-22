package showcash

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Core ...
type Core struct {
	dao   DAO
	useS3 bool
}

// New ...
func New(dao *DAO, useS3 bool) *Core {
	return &Core{
		*dao,
		useS3,
	}
}

func createAuthTokenCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:    "jwt-token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().UTC().Add(time.Hour * 24 * 30),
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Start ...
func (c *Core) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", func(wr http.ResponseWriter, req *http.Request) {
		wr.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	// Setup Context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Static Endpoints
	staticRouter := r.PathPrefix("/static/")
	staticRouter.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../../static"))))

	// API endpoints
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/me", c.apiPostCash).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/me/{guid}", c.apiPutCash).Methods(http.MethodOptions, http.MethodPut)
	apiRouter.HandleFunc("/me/{guid}", c.apiGetCash).Methods(http.MethodOptions, http.MethodGet)

	apiRouter.Use(jsonMiddleware, handlers.CORS(
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodPatch,
		}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Access-Control-Allow-Methods", "Access-Control-Allow-Origin", "Origin", "Accept", "Content-Type"}),
		handlers.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082", "https://api.showcash.io", "https://showcash.io"}),
		handlers.AllowCredentials()),
	)

	http.Handle("/", r)
	log.Println("Showcashing it on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func (c *Core) apiPutCash(wr http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["guid"]
	postID := uuid.FromStringOrNil(slug)

	if postID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	payload := Post{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("apiPutCash.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := c.dao.updatePost(uuid.Nil, payload)
	if err != nil {
		log.Println("updatePost() Failed", err)
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiGetCash(wr http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["guid"]
	postID := uuid.FromStringOrNil(slug)

	if postID == uuid.Nil {
		log.Println("got uuid.Nil")
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	result, err := c.dao.getPost(postID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("getPost() Failed", err)
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiPostCash(wr http.ResponseWriter, req *http.Request) {
	log.Println("Got data")
	payload := struct {
		File     string `json:"file,omitempty"`
		Filename string `json:"filename,omitempty"`
	}{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("apiPostCash Decode() Failed", err)
		return
	}

	log.Println("Uploaded...", payload.Filename)
	dec, err := base64.StdEncoding.DecodeString(payload.File)
	if err != nil {
		log.Println("DecodeString() Failed", err)
		return
	}

	var generatedImageURI string
	if c.useS3 {
		generatedImageURI = ""
	} else {
		generatedImageURI = fmt.Sprintf("http://localhost:8080/static/%s", payload.Filename)
		// Put it local
		f, err := os.Create(fmt.Sprintf("../../static/%s", payload.Filename))
		if err != nil {
			log.Println("Create() failed", err)
			return
		}
		defer f.Close()

		if _, err := f.Write(dec); err != nil {
			log.Println("Write() Failed", err)
			return
		}
		if err := f.Sync(); err != nil {
			log.Println("Sync() Failed", err)
			return
		}
	}

	newPost := Post{
		ImageURI: generatedImageURI,
	}
	result, err := c.dao.createPost(uuid.Nil, newPost)
	if err != nil {
		log.Fatalln("WTF?", err)
	}

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func expiredAuthCookie() *http.Cookie {
	return &http.Cookie{
		Name:    "jwt-token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Now().UTC().Add(-1),
	}
}
