package showcash

import (
	"fmt"
)

var (
	// ErrBadDAO - bad database object
	errBadDAO = fmt.Errorf("ErrBadDAO: Bad Response from DAO (credentials?)")
	// ErrNotImplemented - function not implimented
	errNotImplemented = fmt.Errorf("ErrNotImplemented: Missing function - Probably TBD")
	errNotPresent     = fmt.Errorf("ErrNotPresent: The expected value not set")
	errNotAuthorized  = fmt.Errorf("unauthorized")
)
