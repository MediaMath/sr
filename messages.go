package sr

//Schema is used to post new schemas and as a response to the schema endpoint (uses a string instead of json object for the actual payload)
type Schema struct {
	Schema string `json:"schema"`
}
