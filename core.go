package showcash

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/gofrs/uuid"

	jwt "github.com/dgrijalva/jwt-go"
)

type ShowcashCore struct {
	refreshStrategy RefreshStrategy
}

func New() *ShowcashCore {
	return &ShowcashCore{
		refreshStrategy: DefaultRefreshStrategy(),
	}
}

var indexFile = "../showcash/dist/index.html"

func handlerSPA(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(indexFile); err != nil {
		log.Println("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, indexFile)
}

type RefreshStrategy func(cookie *AuthCookie) (*http.Cookie, error)

func requestWithUserSession(req *http.Request, userSession *AuthCookie) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), struct{}{}, *userSession))
}

func Validate(token string, refreshResultHandler func(*AuthCookie) (*AuthCookie, error)) (*AuthCookie, error) {
	userSession := &AuthCookie{}
	err := ParseTokenWithDefaultSigningKey(token, userSession)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors&(jwt.ValidationErrorExpired) != 0 && validForRefresh(userSession) {
			return refreshResultHandler(userSession)
		}
		return nil, err
	}
	return userSession, nil
}

func createAuthTokenCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:    "jwt-token",
		Value:   token,
		Path:    "/",
		Expires: time.Now().UTC().Add(time.Hour * 24 * 30),
	}
}

func SignedUserToken(email string, userID uuid.UUID, userStatus UserStatus) (string, error) {
	rightNow := time.Now().UTC()
	return SignClaimsWithDefaultSigningKey(AuthCookie{
		Email:      email,
		UserID:     userID,
		UserStatus: userStatus,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rightNow.Add(15 * time.Minute).Unix(),
			IssuedAt:  rightNow.Unix(),
		},
	})
}

func DefaultRefreshStrategy() RefreshStrategy {
	return func(userSession *AuthCookie) (*http.Cookie, error) {
		// user, err := dao.FindUserByID(userSession.UserID)
		// if err != nil {
		// 	return nil, err
		// }
		// if user.UserStatus == UserApproved {
		token, err := SignedUserToken("nick@showcash.io", uuid.Nil, UserApproved)
		if err != nil {
			return nil, err
		}
		return createAuthTokenCookie(token), nil
		// }
		// return nil, errNotAuthorized
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (c *ShowcashCore) apiContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		log.Println("Here")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				log.Println("1")
				if userSession, err := validateAndRefresh(req, wr, c.refreshStrategy); err == nil {
					next.ServeHTTP(wr, requestWithUserSession(req, userSession))
					return
				} else {
					log.Println("Got error:", err)
				}
			}
		}
		next.ServeHTTP(wr, req)
	})
}

func validateAndRefresh(req *http.Request, wr http.ResponseWriter, refreshStrategy RefreshStrategy) (*AuthCookie, error) {
	cookie, err := req.Cookie("jwt-token")
	if err != nil {
		return nil, err
	}
	return Validate(cookie.Value, func(userSession *AuthCookie) (*AuthCookie, error) {
		refreshedCookie, err := refreshStrategy(userSession)
		if err != nil {
			http.SetCookie(wr, expiredAuthCookie())
			return nil, err
		}
		http.SetCookie(wr, refreshedCookie)
		return userSession, nil
	})
}

func (c *ShowcashCore) Start() {
	r := mux.NewRouter()

	// Setup Context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Couldn't be fucked with the static magic
	staticCSSRouter := r.PathPrefix("/css")
	staticJSRouter := r.PathPrefix("/js")
	staticImgRouter := r.PathPrefix("/img")
	staticCSSRouter.Handler(http.StripPrefix("/css", http.FileServer(http.Dir("../showcash/dist/css"))))
	staticJSRouter.Handler(http.StripPrefix("/js", http.FileServer(http.Dir("../showcash/dist/js"))))
	staticImgRouter.Handler(http.StripPrefix("/img", http.FileServer(http.Dir("../showcash/dist/img"))))

	// API Endpoints

	// External webhook and form handler
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.HandleFunc("/me", c.apiMethodTestMe).Methods(http.MethodOptions, http.MethodGet, http.MethodPut)
	apiRouter.HandleFunc("/login", c.apiLogin).Methods(http.MethodGet)
	apiRouter.HandleFunc("/me", corsNop).Methods(http.MethodOptions)

	apiRouter.Use(jsonMiddleware, handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Authorization", "Access-Control-Allow-Methods", "Access-Control-Allow-Origin", "Origin", "Accept", "Content-Type"}),
		handlers.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082", "https://api.showcash.io", "https://showcash.io"}),
		handlers.AllowCredentials()),
		c.apiContextMiddleware)

	r.NotFoundHandler = r.NewRoute().HandlerFunc(handlerSPA).GetHandler()
	http.Handle("/", r)
	log.Println("Doing it....")
	http.ListenAndServe(":8080", nil)
}

func (c *ShowcashCore) apiMethodTestMe(wr http.ResponseWriter, req *http.Request) {
	token, ok := getAuthorisedUserToken(req)
	log.Println("token:", token)
	if !ok {
		JSONRespondWith(wr, apiUnauthorizedError)
		return
	}
	log.Println("Hit", token)
}

func (c *ShowcashCore) apiLogin(wr http.ResponseWriter, req *http.Request) {
	// user, err := c.dao.FindUserByEmail(req.Form.Get("email"))
	// if err == nil && user != nil && quicka.HashMatchesPlaintext(user.PasswordHash, req.Form.Get("password")) && (user.UserStatus == quicka.UserPendingKYCReview || user.UserStatus == quicka.UserApproved) {
	token, err := SignedUserToken("nick@showcash.io", uuid.Nil, UserApproved)
	if err != nil {
		JSONRespondWith(wr, apiServerError)
		return
	}
	http.SetCookie(wr, createAuthTokenCookie(token))
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
