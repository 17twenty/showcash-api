package showcash

import (
	"database/sql/driver"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
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

// AuthCookie - cookie for auth
type AuthCookie struct {
	Email      string
	UserID     uuid.UUID
	UserStatus UserStatus
	jwt.StandardClaims
}

// UserStatus (must be int64)
// is the status of a user
type UserStatus int64

const (
	// UserStatusUnknown is the Database default which means somethign fucked up
	UserStatusUnknown UserStatus = iota
	// UserApproved means all good
	UserApproved
	// UserAccountClosedByUser is as mentioned
	UserAccountClosedByUser
	// UserAccountSuspended due to naughty user
	UserAccountSuspended
)

func (x UserStatus) String() string {
	return map[UserStatus]string{
		UserStatusUnknown:       "UserStatusUnknown",
		UserApproved:            "UserApproved",
		UserAccountClosedByUser: "UserAccountClosedByUser",
		UserAccountSuspended:    "UserAccountSuspended",
	}[x]
}

// Scan ...
func (x *UserStatus) Scan(value interface{}) error {
	db := value.([]uint8)
	*x = map[string]UserStatus{
		"UserStatusUnknown":       UserStatusUnknown,
		"UserApproved":            UserApproved,
		"UserAccountClosedByUser": UserAccountClosedByUser,
		"UserAccountSuspended":    UserAccountSuspended,
	}[string(db)]
	return nil
}

// Value ...
func (x UserStatus) Value() (driver.Value, error) {
	return x.String(), nil
}
