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
