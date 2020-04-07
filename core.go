package showcash

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/17twenty/gorillimiter"
	"github.com/17twenty/showcash-api/pkg/jogly"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var awsSession *session.Session

// Core ...
type Core struct {
	dao   DAO
	useS3 bool
}

// New ...
func New(dao *DAO, useS3 bool) *Core {
	if useS3 {
		var err error
		if awsSession, err = session.NewSession(&aws.Config{
			Region: aws.String("ap-southeast-2")},
		); err != nil {
			log.Panic("Couldn't create AWS session after requesting", err)
		}
	}
	return &Core{
		*dao,
		useS3,
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

	// Auth endpoints
	authRouter := r.PathPrefix("/auth/").Subrouter()
	authRouter.HandleFunc("/login", c.apiPostLogin).Methods(http.MethodOptions, http.MethodPost)
	authRouter.HandleFunc("/logout", c.apiGetLogout).Methods(http.MethodOptions, http.MethodGet)
	authRouter.HandleFunc("/register", c.apiPostSignup).Methods(http.MethodOptions, http.MethodPost)

	// API endpoints
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/view", c.apiPostIncreaseView).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/mostviewed", c.apiGetMostViewed).Methods(http.MethodOptions, http.MethodGet)
	apiRouter.HandleFunc("/recent", c.apiGetMostRecent).Methods(http.MethodOptions, http.MethodGet)
	apiRouter.HandleFunc("/comments/{guid}", c.apiGetComments).Methods(http.MethodOptions, http.MethodGet)
	apiRouter.HandleFunc("/comments/{guid}", authMiddleware(c.apiPostComment)).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/me", authMiddleware(c.apiPostCash)).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/remove/{guid}", c.apiDeletePost).Methods(http.MethodOptions, http.MethodDelete)
	apiRouter.HandleFunc("/claim/{uuid}/{guid}", c.apiClaimPost).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/me/{guid}", authMiddleware(c.apiPutCash)).Methods(http.MethodOptions, http.MethodPut)
	apiRouter.HandleFunc("/me/{guid}", c.apiGetCash).Methods(http.MethodOptions, http.MethodGet)
	apiRouter.HandleFunc("/profile", authMiddleware(c.apiGetMe)).Methods(http.MethodOptions, http.MethodGet)
	apiRouter.HandleFunc("/profile", authMiddleware(c.apiPutMe)).Methods(http.MethodOptions, http.MethodPut)
	apiRouter.HandleFunc("/profile/{handle}", c.apiGetUserProfile).Methods(http.MethodOptions, http.MethodGet)

	// Waitlist goes to Slack
	apiRouter.HandleFunc("/waitlist", c.apiPostWaitlist).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/recommend", c.apiPostRecommend).Methods(http.MethodOptions, http.MethodPost)

	defaultLimiter := func(next http.Handler) http.Handler {
		return gorillimiter.Limiter(next, 3, time.Second)
	}
	authRouter.Use(jsonMiddleware,
		defaultLimiter,
		handlers.CORS(
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
			handlers.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081", "https://api.showcash.io", "https://showcash.io"}),
			handlers.AllowCredentials()),
	)

	apiRouter.Use(jsonMiddleware,
		defaultLimiter,
		handlers.CORS(
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
			handlers.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081", "https://api.showcash.io", "https://showcash.io"}),
			handlers.AllowCredentials()),
	)

	http.Handle("/", r)
	log.Println("Showcashing on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func (c *Core) apiPostRecommend(wr http.ResponseWriter, req *http.Request) {
	v := struct {
		Name         string `json:"name,omitempty"`
		EmailAddress string `json:"email_address,omitempty"`
		Why          string `json:"why,omitempty"`
	}{}
	if err := json.NewDecoder(req.Body).Decode(&v); err != nil {
		log.Println("apiPostComment.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	jogly.New("https://api.jogly.io/e/hook/0a5a6809-d96f-4ade-a0eb-903c22dedb3a/").Serialise(struct {
		Name         string
		EmailAddress string
		Why          string
	}{
		Name:         v.Name,
		EmailAddress: v.EmailAddress,
		Why:          v.Why,
	}).Post()
}

func (c *Core) apiPostWaitlist(wr http.ResponseWriter, req *http.Request) {
	// https://api.jogly.io/e/hook/0a5a6809-d96f-4ade-a0eb-903c22dedb3a/
	// Post to Showcash
	u := User{}
	if err := json.NewDecoder(req.Body).Decode(&u); err != nil {
		log.Println("apiPostWaitlist.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	jogly.New("https://api.jogly.io/e/hook/0a5a6809-d96f-4ade-a0eb-903c22dedb3a/").Serialise(struct {
		Username     string
		Location     string
		Realname     string
		Bio          string
		EmailAddress string
	}{
		Username:     u.Username,
		Location:     u.Location,
		Realname:     u.RealName,
		Bio:          u.Bio,
		EmailAddress: u.EmailAddress,
	}).Post()
}

func (c *Core) apiPostComment(wr http.ResponseWriter, req *http.Request) {
	postID := uuid.FromStringOrNil(mux.Vars(req)["guid"])
	if postID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	u := GetSessionFromContext(req)
	if u == nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	comment := Comment{}
	if err := json.NewDecoder(req.Body).Decode(&comment); err != nil {
		log.Println("apiPostComment.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	comment.Username = u.Username
	comment.UserID = u.UserID
	result, err := c.dao.createComment(u.UserID, postID, comment)
	if err != nil {
		log.Println("apiPostComment().createComment failed", err)
	}
	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiGetMe(wr http.ResponseWriter, req *http.Request) {
	u := GetSessionFromContext(req)
	if u != nil {
		profile, _ := c.dao.getUserProfileByID(u.UserID)
		if err := json.NewEncoder(wr).Encode(profile); err != nil {
			log.Printf("Error Encoding JSON: %s", err)
		}
	}
}
func (c *Core) apiPutMe(wr http.ResponseWriter, req *http.Request) {
	session := GetSessionFromContext(req)
	if session != nil {

		user := User{}
		// Get the payload
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			log.Println("apiPutMe.Decode() failed", err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Verify we only modify the logged in user
		user.UserID = session.UserID

		current, _ := c.dao.updateUser(user)
		if err := json.NewEncoder(wr).Encode(current); err != nil {
			log.Printf("Error Encoding JSON: %s", err)
		}
	}
}
func (c *Core) apiGetUserProfile(wr http.ResponseWriter, req *http.Request) {
	handle := mux.Vars(req)["handle"]
	if !isValidHandle(handle) {
		wr.WriteHeader(http.StatusNotFound)
		return
	}
	user, _ := c.dao.getUserProfileByHandle(handle)
	user.Friends = []UserProfile{} // remove user friends
	if err := json.NewEncoder(wr).Encode(user); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiGetComments(wr http.ResponseWriter, req *http.Request) {
	postID := uuid.FromStringOrNil(mux.Vars(req)["guid"])
	if postID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	result := c.dao.getCommentsForPostID(postID)

	n := 0
	for _, ip := range result {
		if !isAllowed(ip.Comment) {
			continue
		}

		result[n] = ip
		n++
	}

	result = result[:n]

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiPostIncreaseView(wr http.ResponseWriter, req *http.Request) {
	payload := struct {
		ID         uuid.UUID `json:"id,omitempty"`
		Identifier string    `json:"identifier,omitempty"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("apiIncreaseView.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.dao.increaseView(payload.ID, payload.Identifier)
}
func (c *Core) apiClaimPost(wr http.ResponseWriter, req *http.Request) {
	postID := uuid.FromStringOrNil(mux.Vars(req)["guid"])
	userID := uuid.FromStringOrNil(mux.Vars(req)["uuid"])

	if userID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}
	msg := "ok"
	err := c.dao.claimPost(userID, postID)
	if err != nil {
		msg = fmt.Sprintf("Error %v", err)
	}
	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(struct {
		Result string `json:"result,omitempty"`
	}{
		Result: msg,
	}); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}
func (c *Core) apiDeletePost(wr http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["guid"]
	postID := uuid.FromStringOrNil(slug)

	if postID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}
	c.dao.deletePost(postID)
	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(struct {
		Result string `json:"result,omitempty"`
	}{
		Result: "ok",
	}); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}

}

func (c *Core) apiPutCash(wr http.ResponseWriter, req *http.Request) {
	slug, _ := mux.Vars(req)["guid"]
	postID := uuid.FromStringOrNil(slug)

	if postID == uuid.Nil {
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	u := GetSessionFromContext(req)
	if u == nil {
		log.Println("No user context")
		wr.WriteHeader(http.StatusNotFound)
		return
	}

	payload := Post{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("apiPutCash.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := c.dao.updatePost(u.UserID, payload)
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

func (c *Core) apiGetMostRecent(wr http.ResponseWriter, req *http.Request) {
	result := c.dao.getLatestPosts()
	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiGetMostViewed(wr http.ResponseWriter, req *http.Request) {
	result := c.dao.getMostViewedPosts()
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

	n := 0
	for _, ip := range result.ItemList {
		if !isAllowed(ip.Link) || !isAllowed(ip.Description) {
			continue
		}

		result.ItemList[n] = ip
		n++
	}

	result.ItemList = result.ItemList[:n]

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(result); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *Core) apiPostCash(wr http.ResponseWriter, req *http.Request) {
	u := GetSessionFromContext(req)
	if u == nil {
		log.Println("Invalid post", u.Username)
		wr.WriteHeader(http.StatusNotFound)
		return
	}
	payload := struct {
		File     string `json:"file,omitempty"`
		Filename string `json:"filename,omitempty"`
	}{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("apiPostCash Decode() Failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Uploaded...", payload.Filename)
	dec, err := base64.StdEncoding.DecodeString(payload.File)
	if err != nil {
		log.Println("DecodeString() Failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	var generatedImageURI string
	if c.useS3 {
		suffix := filepath.Ext(payload.Filename)
		fileName := fmt.Sprintf("%s%s", uuid.Must(uuid.NewV4()), suffix)
		uploader := s3manager.NewUploader(awsSession)
		fileType := http.DetectContentType(dec)

		resp, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String("showcash-uploads"),
			Key:         aws.String(fileName),
			Body:        bytes.NewReader(dec),
			ContentType: aws.String(fileType),
		})
		if err != nil {
			log.Println("s3 upload() Failed", err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("Uploaded to:", resp.Location)
		generatedImageURI = fmt.Sprintf("https://images.showcash.io/%s", fileName)

	} else {
		generatedImageURI = fmt.Sprintf("http://localhost:8080/static/%s", payload.Filename)
		// Put it local
		f, err := os.Create(fmt.Sprintf("../../static/%s", payload.Filename))
		if err != nil {
			log.Println("Create() failed", err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := f.Write(dec); err != nil {
			log.Println("Write() Failed", err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := f.Sync(); err != nil {
			log.Println("Sync() Failed", err)
			wr.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	newPost := Post{
		ImageURI: generatedImageURI,
	}
	result, err := c.dao.createPost(u.UserID, newPost)
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
