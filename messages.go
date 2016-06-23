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

//Version is a schema plus subject and version information
type Version struct {
	Schema  string `json:"schema"`
	Version int    `json:"version"`
	Subject string `json:"subject"`
}

//CheckedSchema is a schema plus subject, id and version information
type CheckedSchema struct {
	Schema  string `json:"schema"`
	Version int    `json:"version"`
	Subject string `json:"subject"`
	ID      int    `json:"id"`
}

//SchemaID is returned when creating a schema
type SchemaID struct {
	ID int `json:"id"`
}

//IsCompatible is returned on compatibility checks
type IsCompatible struct {
	IsCompatible bool `json:"is_compatible"`
}
