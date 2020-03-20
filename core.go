package showcash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type ShowcashCore struct {
}

func New() *ShowcashCore {
	return &ShowcashCore{}
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

func (c *ShowcashCore) Start() {
	r := mux.NewRouter()

	// Setup Context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// API Endpoints

	// External webhook and form handler
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/me", c.apiMethodTestMe).Methods(http.MethodOptions, http.MethodGet, http.MethodPut)
	apiRouter.HandleFunc("/login", c.apiLogin).Methods(http.MethodGet)

	apiRouter.Use(jsonMiddleware, handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Access-Control-Allow-Methods", "Access-Control-Allow-Origin", "Origin", "Accept", "Content-Type"}),
		handlers.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082", "https://api.showcash.io", "https://showcash.io"}),
		handlers.AllowCredentials()),
	)

	http.Handle("/", r)
	log.Println("Doing it....")
	http.ListenAndServe(":8080", nil)
}

func (c *ShowcashCore) apiMethodTestMe(wr http.ResponseWriter, req *http.Request) {
	log.Println("Got data")

	req.ParseMultipartForm(32 << 20) // limit your max input length!
	var buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := req.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	fmt.Printf("File name %s\n", name[0])
	// Copy the file data to my buffer
	io.Copy(&buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example
	contents := buf.String()
	fmt.Println(contents)
	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	buf.Reset()
	// do something else
	// etc write header
	return

	// 	dec, err := base64.StdEncoding.DecodeString(b64)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	f, err := os.Create("myfilename")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer f.Close()

	// 	if _, err := f.Write(dec); err != nil {
	// 		panic(err)
	// 	}
	// 	if err := f.Sync(); err != nil {
	// 		panic(err)
	// 	}
	// }
}

func (c *ShowcashCore) apiLogin(wr http.ResponseWriter, req *http.Request) {
	// user, err := c.dao.FindUserByEmail(req.Form.Get("email"))
	// if err == nil && user != nil && quicka.HashMatchesPlaintext(user.PasswordHash, req.Form.Get("password")) && (user.UserStatus == quicka.UserPendingKYCReview || user.UserStatus == quicka.UserApproved) {

	log.Println("This is shit")
	// token, err := SignedUserToken("nick@showcash.io", uuid.Nil, UserApproved)
	// if err != nil {
	// 	JSONRespondWith(wr, apiServerError)
	// 	return
	// }
	// http.SetCookie(wr, createAuthTokenCookie(token))
	// }
}

// GetAuthorisedUserToken -
func getAuthorisedUserToken(req *http.Request) (AuthCookie, bool) {
	ctx := req.Context().Value(struct{}{})
	val := AuthCookie{}
	var ok bool
	if val, ok = ctx.(AuthCookie); ok {
		return val, true
	}
	return val, false
}

func corsNop(wr http.ResponseWriter, req *http.Request) {
	JSONRespondWith(wr, apiOK)
}

// JSONRespondWith - handles JSON response with Status code
func JSONRespondWith(wr http.ResponseWriter, resp apiSimpleResponse) {
	wr.Header().Set("Content-Type", "application/json")
	wr.WriteHeader(resp.statusCode)
	j := json.NewEncoder(wr)
	if err := j.Encode(resp); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func validForRefresh(userSession *AuthCookie) bool {
	return time.Now().Before(time.Unix(userSession.ExpiresAt, 0).Add(14 * 24 * time.Hour))
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
