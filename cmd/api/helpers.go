package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// This doesn't use any dependency from `application`, but it's good
// practice anyway to maintain consistency in code structure.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// Interpolated URL paramters are stored in the request context
	params := httprouter.ParamsFromContext(r.Context())

	// Convert the string value (returned by ByName) of the ID into base 10 integer
	// (with a bit size of 64 --  how many bits the result should fit into [64 = int64]).
	// If unable to parse -- or ID is less than 1, it is
	// invalid, so throw notFound
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id paramter")
	}

	return id, nil
}
