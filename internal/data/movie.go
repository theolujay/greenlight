package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

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
