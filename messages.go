package sr

import (
	"encoding/json"
	"fmt"
)

//Error is returned on errors
type Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%s", bytes)
}

//Schema is used to post new schemas and as a response to the schema endpoint (uses a string instead of json object for the actual payload)
type Schema struct {
	Schema string `json:"schema"`
}

//SchemaID is returned when creating a schema
type SchemaID struct {
	ID int `json:"id"`
}
