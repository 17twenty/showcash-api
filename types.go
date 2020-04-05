package showcash

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

var (
	apiServerError       = apiSimpleResponse{Error: "Generic Server Error", statusCode: http.StatusInternalServerError}
	apiBadRequestError   = apiSimpleResponse{Error: "Bad Request", statusCode: http.StatusBadRequest}
	apiUnauthorizedError = apiSimpleResponse{Error: "Unauthorized", statusCode: http.StatusUnauthorized}
	apiNotFoundError     = apiSimpleResponse{Error: "Not Found", statusCode: http.StatusNotFound}
	apiOK                = apiSimpleResponse{Message: "success", statusCode: http.StatusOK}
)

type apiSimpleResponse struct {
	Error      string   `json:"error,omitempty"`
	Message    string   `json:"message,omitempty"`
	Messages   []string `json:"messages,omitempty"`
	statusCode int
}

// Item is the dope things
type Item struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Left        int    `json:"left"`
	Top         int    `json:"top"`
}

// Post is the type used for wrapping cool shit
type Post struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	ImageURI string    `json:"imageuri"`
	Date     time.Time `json:"date"`
	ItemList []Item    `json:"itemList"`
}

// Comment is a comment posted on a post
type Comment struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Date     time.Time `json:"date,omitempty"`
	Comment  string    `json:"comment,omitempty"`
	Username string    `json:"username,omitempty"`
	UserID   uuid.UUID `json:"user_id,omitempty"`
	// Points   int       `json:"points"`    // How many points this comment has
	// HasVoted int       `json:"has_voted"` // If the user voted it up or down -1 | 0 | 1
}

// User is a showcash user
type User struct {
	UserID       uuid.UUID `json:"user_id,omitempty"`
	Username     string    `json:"username,omitempty"`
	RealName     string    `json:"realname,omitempty"`
	Location     string    `json:"location,omitempty"`
	ProfileURI   string    `json:"profile_uri,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	Social1      string    `json:"social_1,omitempty"`
	Social2      string    `json:"social_2,omitempty"`
	Social3      string    `json:"social_3,omitempty"`
	EmailAddress string    `json:"email_address,omitempty"`
	Password     string    `json:"password,omitempty"`
	ShadowBanned bool      `json:"shadow_banned,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}
