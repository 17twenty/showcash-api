package showcash

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofrs/uuid"
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

	// Static Endpoints
	staticRouter := r.PathPrefix("/static/")
	staticRouter.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../../static"))))

	// API endpoints
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/me", c.apiPostCash).Methods(http.MethodOptions, http.MethodPost)
	apiRouter.HandleFunc("/me/{slug}", c.apiPutCash).Methods(http.MethodOptions, http.MethodPut)
	apiRouter.HandleFunc("/me/{slug}", c.apiGetCash).Methods(http.MethodOptions, http.MethodGet)

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

// Item is the dope things
type Item struct {
	ID          int    `json:"id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Link        string `json:"link,omitempty"`
	Left        int    `json:"left,omitempty"`
	Top         int    `json:"top,omitempty"`
}

// ShowCash is the type used for wrapping cool shit
type ShowCash struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Title    string    `json:"title,omitempty"`
	ImageURI string    `json:"imageURI,omitempty"`
	Date     time.Time `json:"date,omitempty"`
	ItemList []Item    `json:"itemList,omitempty"`
}

func (c *ShowcashCore) apiPutCash(wr http.ResponseWriter, req *http.Request) {
	// TODO: Use the slug!
	slug, _ := mux.Vars(req)["guid"]
	log.Println("Requested", slug)

	payload := ShowCash{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("Post external invoice API: Unable to decode JSON request", err)
		return
	}

	log.Println("Got:", payload)
}

func (c *ShowcashCore) apiGetCash(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(ShowCash{
		ID:       uuid.FromStringOrNil("92201c6c-0929-42e4-ae30-58436ba80419"),
		Title:    "The baddest tequila",
		ImageURI: fmt.Sprintf("http://localhost:8080/static/%s", "Overview.png"),
		Date:     time.Now(),
		ItemList: []Item{
			{
				ID:          0,
				Left:        80,
				Top:         80,
				Title:       "My Shit",
				Description: "Boogie woogie",
				Link:        "https://www.google.com",
			},
		},
	}); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *ShowcashCore) apiPostCash(wr http.ResponseWriter, req *http.Request) {
	log.Println("Got data")
	payload := struct {
		File     string `json:"file,omitempty"`
		Filename string `json:"filename,omitempty"`
	}{}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		log.Println("Post external invoice API: Unable to decode JSON request", err)
		return
	}

	log.Println("Uploaded...", payload.Filename)
	dec, err := base64.StdEncoding.DecodeString(payload.File)
	if err != nil {
		log.Println("WTF1", err)
		return
	}

	f, err := os.Create(fmt.Sprintf("../../static/%s", payload.Filename))
	if err != nil {
		log.Println("WTF2", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(dec); err != nil {
		log.Println("WTF3", err)
		return
	}
	if err := f.Sync(); err != nil {
		log.Println("WTF4", err)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(wr).Encode(ShowCash{
		ImageURI: fmt.Sprintf("http://localhost:8080/static/%s", payload.Filename),
		ID:       uuid.FromStringOrNil("92201c6c-0929-42e4-ae30-58436ba80419"), //(uuid.NewV4()),
	}); err != nil {
		log.Printf("Error Encoding JSON: %s", err)
	}
}

func (c *ShowcashCore) apiLogin(wr http.ResponseWriter, req *http.Request) {
	log.Println("called apiLogin")
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
