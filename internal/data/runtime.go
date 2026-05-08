package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Runtime.MarshalJSON is a custom method that json.Marshal() will call in
// order to specifically encode the Movie.Runtime field with its time unit.
//
// While this is convenient and nice to use, the downside is that it can
// sometimes be awkward when integrating code with other packages, and one
// may need to perform type conversions to change the custom type to and
// from a value that the other packages understand and accept.
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a stirng containing the movie runtime in the required format.
	jsonValue := fmt.Sprintf("%d mins", r)
	// Because jsonValue is a string value, it must be wrapped in double quotes
	// before returned; otherwise it won't be interpreted as a JSON string,
	// resulting in a runtime error
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// The UnmarshalJSON Runtime method satisfies the json.Unmarshaler interface
// for json.Decode to call instead to specifically parse the runtime field
// in a custom manner. It is essentially the reverse of Runtime.MarshalJSON()
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// It is expected that jsonValue will be a string in the format
	// "<runtime> mins", so we unquote it and return ErrInvalidRuntimeFormat
	// error if unable to.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	ok := strings.HasSuffix(unquotedJSONValue, " mins")
	if !ok {
		return ErrInvalidRuntimeFormat
	}
	fieldValue := strings.TrimSuffix(unquotedJSONValue, " mins")
	i, err := strconv.ParseInt(fieldValue, 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(i)

	return nil
}
