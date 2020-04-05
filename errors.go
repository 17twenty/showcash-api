package showcash

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/lib/pq"
)

var (
	// ErrBadDAO - bad database object
	errBadDAO = fmt.Errorf("ErrBadDAO: Bad Response from DAO (credentials?)")
	// ErrNotImplemented - function not implimented
	errNotImplemented = fmt.Errorf("ErrNotImplemented: Missing function - Probably TBD")
	errNotPresent     = fmt.Errorf("ErrNotPresent: The expected value not set")
	errNotAuthorized  = fmt.Errorf("unauthorized")
)

func jsonResponse(wr http.ResponseWriter, message string, code int) {
	wr.WriteHeader(code)
	if err := json.NewEncoder(wr).Encode(struct {
		Message string
	}{
		Message: message,
	}); err != nil {
		log.Println("Failed to write err:", err)
	}
}

// Database errors as needed
var (
	errNotUnique = errors.New("Violated a unique constraint")
)

func pgErrIs(got error, test error) bool {
	if got == nil {
		return false
	}
	if std := errors.Is(got, test); std {
		return true
	}
	if err, ok := got.(*pq.Error); ok {
		switch e := err.Code; e {
		case "23505":
			if test == errNotUnique {
				return true
			}
		default:
			log.Println("Unknown Code", e)
		}
	} else {
		log.Println("Not OK -- not the right error")
	}

	return false
}
