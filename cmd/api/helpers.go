package main

import (
	// "maps"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

// readIDParam pulls the value of `id` in a URL path, like in
// <baseURL>/movies/:id
//
// It doesn't use any dependency from `application`, but it's good
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

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, returning the error if there's one
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Add a newline to make it easier to view in the terminal
	js = append(js, '\n')

	// There surely won't be any more errors at this point before writing
	// the response, so it's safe to add a any headers to be included. It's
	// OK if the privided header map is nil. Go doesn't throw and error if
	// it is to range over (or generally, read from) a nil map, even though
	// using maps.Copy is also possible cleaner.
	// maps.Copy(w.Header(), headers)
	for key, val := range headers {
		w.Header()[key] = val
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
