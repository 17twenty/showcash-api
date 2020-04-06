package showcash

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"

	"github.com/gofrs/uuid"
	"github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte("NdRgUkXp2r5u8x/A?D(G+KbPeShVmYq3")
	blockKey = []byte("Xn2r5u8x/A?D(G+KbPeShVmYp3s6v9y$")
)

var sc = securecookie.New(hashKey, blockKey)

func isValidHandle(s string) bool {
	if len(s) < 2 || len(s) > 16 {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') &&
			(r != '_') &&
			(r != '-') {
			return false
		}
	}
	return true
}

func (c *Core) apiPostLogin(wr http.ResponseWriter, req *http.Request) {
	v := struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&v); err != nil {
		log.Println("apiPostLogin.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isValidHandle(v.Username) {
		jsonResponse(wr, "Bad Creds", http.StatusForbidden)
		return
	}

	if user, err := c.dao.getUserByUsernameAndPassword(v.Username, v.Password); err == nil {
		setUserCookie(wr, user)
		user.Password = ""
		user.ShadowBanned = false
		if err := json.NewEncoder(wr).Encode(user); err != nil {
			log.Printf("Error Encoding JSON: %s", err)
		}
		return
	} else if pgErrIs(err, sql.ErrNoRows) {
		log.Println("No such user")
	} else {
		log.Println("Got a weird error:", err)
	}

	jsonResponse(wr, "Bad Creds", http.StatusForbidden)
}

func setUserCookie(wr http.ResponseWriter, u User) {
	if encoded, err := sc.Encode("showcash", struct {
		UserID       uuid.UUID
		Username     string
		EmailAddress string
	}{
		u.UserID,
		u.Username,
		u.EmailAddress,
	}); err == nil {
		cookie := &http.Cookie{
			Name:  "showcash",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(wr, cookie)
	}
}

func clearUserCookie(wr http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "showcash",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(wr, cookie)
}

func authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		if cookie, err := req.Cookie("showcash"); err == nil {
			u := User{}
			if err = sc.Decode("showcash", cookie.Value, &u); err == nil {
				// log.Println("Just saw:", u.Username, u.UserID, u.EmailAddress)
				h.ServeHTTP(wr, RequestWithUserSession(req, u)) // call ServeHTTP on the original handler
				return
			}
		}
		jsonResponse(wr, "Couldnt get the right stuff", http.StatusForbidden)
	})
}

// RequestWithUserSession will create a new request and will attach the userSession
// in the context of the request
func RequestWithUserSession(req *http.Request, user User) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), struct{}{}, user))
}

// GetSessionFromContext will take the authorized user session shared on the
// http request
func GetSessionFromContext(req *http.Request) *User {
	ctx := req.Context().Value(struct{}{})
	n, ok := ctx.(User)
	if ok {
		return &n
	}
	return nil
}

func (c *Core) apiGetLogout(wr http.ResponseWriter, req *http.Request) {
	u := GetSessionFromContext(req)
	if u != nil {
		log.Println("Just logged out", u.Username)
	}
	clearUserCookie(wr)
}

func (c *Core) apiPostSignup(wr http.ResponseWriter, req *http.Request) {
	newUser := User{}
	if err := json.NewDecoder(req.Body).Decode(&newUser); err != nil {
		log.Println("apiPostSignup.Decode() failed", err)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Email check
	_, err := mail.ParseAddress(newUser.EmailAddress)
	if err != nil {
		jsonResponse(wr, "Garbage email", http.StatusBadRequest)
		return
	}

	// Handle check
	if !isValidHandle(newUser.Username) || !isAllowed(newUser.RealName) {
		jsonResponse(wr, "Name too short or just rude", http.StatusBadRequest)
		return
	}

	result, err := c.dao.createUser(newUser)
	if pgErrIs(err, errNotUnique) {
		jsonResponse(wr, "Username or email exists... do you have an account?", http.StatusConflict)
		return
	} else if err != nil {
		log.Println("err:", err)
	}

	// Pass to login
	setUserCookie(wr, result)
}
