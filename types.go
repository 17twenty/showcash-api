package showcash

import (
	"net/http"
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
